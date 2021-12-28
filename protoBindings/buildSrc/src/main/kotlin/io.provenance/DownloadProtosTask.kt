package io.provenance

import org.apache.commons.compress.archivers.tar.TarArchiveEntry
import org.apache.commons.compress.archivers.tar.TarArchiveInputStream
import org.apache.commons.io.FileUtils
import org.apache.commons.io.IOUtils
import org.apache.http.client.methods.HttpGet
import org.apache.http.impl.client.HttpClients
import org.apache.http.impl.client.LaxRedirectStrategy
import org.gradle.api.DefaultTask
import org.gradle.api.tasks.Input
import org.gradle.api.tasks.TaskAction
import org.gradle.api.tasks.options.Option
import java.io.File
import java.io.FileInputStream
import java.io.FileOutputStream
import java.io.IOException
import java.io.InputStream
import java.nio.file.Paths
import java.util.zip.GZIPInputStream
import java.util.zip.ZipEntry
import java.util.zip.ZipFile

/**
 * Custom gradle task to download Provenance and Cosmos protobuf files.
 */
open class DownloadProtosTask : DefaultTask() {
    private val tempPrefix = this.javaClass.name

    @Option(
        option = "cosmos-version",
        description = "Cosmos release version (e.g. v0.44.3)"
    )
    @Input
    var cosmosVersion: String? = null

    @Option(
        option = "wasmd-version",
        description = "provenance-io/wasmd release version (e.g. v0.19.0)"
    )
    @Input
    var wasmdVersion: String? = null

    @Option(
        option = "ibc-version",
        description = "Cosmos IBC release version (e.g. v1.1.0)"
    )
    @Input
    var ibcVersion: String? = null

    /**
     * Hard coded to v0.17.0 for backwards compatibility. Needed to serialize/deserialize older wasmd protos.
     */
    @Input
    val cosmwasmVersion: String = "v0.17.0"

    /**
     * Connects directly to provenance-io GitHub release directory
     * and downloads the `provenanceVersion` proto zip file.
     *
     * Connects directly to cosmos-sdk GitHub tarball release directory
     * and downloads the `cosmosVersion` proto gzipped tar file.
     *
     * Connects directly to provenance-io/wasmd GitHub tarball release directory
     * and downloads the `wasmdVersion` proto gzipped tar file.
     *
     * All files are uncompressed into the `third_party/proto` directory
     * of this root gradle project.
     *
     */
    @TaskAction
    fun downloadProtos() {

        cleanDestination(thirdPartyPath())

        untar(
            file = unGzip(toTempFile("https://github.com/provenance-io/wasmd/tarball/${this.wasmdVersion}")),
            destinationDir = thirdPartyPath(),
            includePattern = Regex(".*/proto/.*\\.proto\$"),
            excludePattern = Regex(".*third_party/.*|.*proto/ibc/.*"),
            protoRootDir = "proto"
        )

        untar(
            file = unGzip(toTempFile("https://github.com/cosmos/ibc-go/tarball/${this.ibcVersion}")),
            destinationDir = thirdPartyPath(),
            includePattern = Regex(".*/proto/.*\\.proto\$"),
            excludePattern = Regex("FAILSAFE_STRING"),
            protoRootDir = "proto"
        )

        untar(
            file = unGzip(toTempFile("https://github.com/cosmos/cosmos-sdk/tarball/${this.cosmosVersion}")),
            destinationDir = thirdPartyPath(),
            includePattern = Regex(".*/proto/.*\\.proto\$"),
            excludePattern = Regex(".*testutil/.*"),
            protoRootDir = "proto"
        )

        untar(
            file = unGzip(toTempFile("https://github.com/CosmWasm/wasmd/tarball/${this.cosmwasmVersion}")),
            destinationDir = thirdPartyPath(),
            includePattern = Regex(".*/proto/.*\\.proto\$"),
            excludePattern = Regex(".*third_party/.*|.*proto/ibc/.*"),
            protoRootDir = "proto"
        )
    }

    /**
     * The default destination directory for downloaded and uncompressed
     * protos
     */
    private fun thirdPartyPath() =
        Paths.get(project.rootProject.rootDir.toString(), "build", "third_party").toString()

    /**
     * Clean the destination for the downloaded and uncompressed *.proto
     * files (i.e. `third_party/proto`)
     *
     */
    private fun cleanDestination(destinationDir: String) {
        FileUtils.deleteQuietly(File(destinationDir))
        FileUtils.forceMkdir(File(destinationDir))
    }

    /**
     * Extract `url` to a local machine temp directory
     */
    private fun toTempFile(url: String): File =
        HttpClients.custom().setRedirectStrategy(LaxRedirectStrategy()).build()
            .use { client ->
                client.execute(
                    HttpGet(url)
                ) { response ->
                    if (response == null || response.statusLine.statusCode != 200) {
                        throw IOException("could not retrieve: ${response?.statusLine?.reasonPhrase}")
                    }
                    File.createTempFile(tempPrefix, "zip").let { tempFile ->
                        IOUtils.copy(
                            response.entity.content,
                            FileOutputStream(
                                tempFile
                            )
                        )
                        tempFile
                    }
                }
            }

    /**
     * Unzip the given `file` to `destinationDir` but only include files
     * that match the `includePattern` regex
     */
    private fun unzip(
        file: File,
        destinationDir: String,
        includePattern: Regex
    ) {
        ZipFile(file).use { zip ->
            zip.entries().asSequence()
                .forEach { zipEntry ->
                    handleZipEntry(
                        zipInputStream = zip.getInputStream(zipEntry),
                        zipEntry = zipEntry,
                        destinationDir = File(destinationDir),
                        includePattern = includePattern
                    )
                }
        }
    }

    /**
     * Given a zip input stream and an entry in the zip file, extract
     * the zip entry to the `destinationDir` when the zip entry file matches
     * the `includePattern`
     */
    @Throws(IOException::class)
    private fun handleZipEntry(
        zipInputStream: InputStream,
        zipEntry: ZipEntry,
        destinationDir: File,
        includePattern: Regex
    ) {
        if (zipEntry.name.matches(includePattern)) {
            val newFile = File(destinationDir, zipEntry.name)
            if (zipEntry.isDirectory) {
                if (!newFile.isDirectory && !newFile.mkdirs()) {
                    throw IOException("Failed to create directory $newFile")
                }
            } else {
                // fix for Windows-created archives
                val parent = newFile.parentFile
                if (!parent.isDirectory && !parent.mkdirs()) {
                    throw IOException("Failed to create directory $parent")
                }
                IOUtils.copy(zipInputStream, FileOutputStream(newFile))
            }
        }
    }

    /**
     * UnTar the given `file` to `destinationDir` but only include files
     * that match the `includePattern` regex and don't match the `excludePattern`.
     *
     * The `protoRootDir` is used to find the first occurrence directory of
     * the `proto` directory (for example).  This `protoRootDir` is the directory
     * copied to the local `thirdPartyPath()`.
     */
    @Throws(IOException::class)
    private fun untar(
        file: File,
        destinationDir: String,
        includePattern: Regex,
        excludePattern: Regex,
        protoRootDir: String
    ) {
        val tempDir = File.createTempFile(tempPrefix, "dir").parentFile

        // Keep the first (top) occurrence of a directory in the tar so
        // copying entire directories is simpler
        var topTarDirectory: File? = null

        TarArchiveInputStream(FileInputStream(file)).use { tarArchiveInputStream ->
            var tarEntry: TarArchiveEntry?
            while (tarArchiveInputStream.nextTarEntry.also {
                tarEntry = it
            } != null
            ) {
                if (topTarDirectory == null) {
                    topTarDirectory = File(tempDir.absolutePath + File.separator + tarEntry?.name)
                }

                if (tarEntry?.name?.matches(includePattern) == true &&
                    tarEntry?.name?.matches(excludePattern) == false
                ) {
                    // write to temp file first so we can pick the dirs we want
                    val outputFile = File(tempDir.absolutePath + File.separator + tarEntry?.name)
                    if (tarEntry?.isDirectory == true) {
                        if (!outputFile.exists()) {
                            outputFile.mkdirs()
                        }
                    } else {
                        outputFile.let {
                            it.parentFile.mkdirs()
                            IOUtils.copy(
                                tarArchiveInputStream,
                                FileOutputStream(it)
                            )
                        }
                    }
                }
            }
        }
        // Copy from proto root dir to the local project third_party dir
        topTarDirectory?.let { topTar ->
            mutableListOf<File>().let { matchedDirs ->
                findDirectory(topTar, protoRootDir, matchedDirs)
                matchedDirs.forEach {
                    FileUtils.copyDirectory(it, File("$destinationDir${File.separator}proto"))
                }
            }
        } ?: throw IOException("tar file ${file.absolutePath} is not a well formed tar file - missing top level directory")
    }

    /**
     * Given a cwd find the all of the first directories matching the `findDirectory` name.
     * For example, given a `findDirectory` of `proto` this will return matching
     * directories named:
     * `./some/dir/level/proto/messages` and `./some/other/dir/proto/messages`
     *
     */
    private fun findDirectory(
        currentDirectory: File,
        findDirectory: String,
        matchingDirectories: MutableList<File>
    ) {
        if (currentDirectory.isDirectory && currentDirectory.name == findDirectory) {
            matchingDirectories.add(currentDirectory)
        } else {
            val files = currentDirectory.listFiles() ?: emptyArray()
            for (file in files) {
                if (file.isFile) {
                    continue
                }
                if (file.isDirectory) {
                    findDirectory(file, findDirectory, matchingDirectories)
                }
            }
        }
    }

    /**
     * ungzip a given gZippedFile tar file
     */
    @Throws(IOException::class)
    private fun unGzip(gZippedFile: File): File =
        GZIPInputStream(FileInputStream(gZippedFile)).let { gzip ->
            File.createTempFile(tempPrefix, "tar").let { tempFile ->
                FileUtils.copyToFile(gzip, tempFile)
                tempFile
            }
        }
}
