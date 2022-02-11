import org.gradle.api.Project

// Check to see if a version number is passed via `-PartifactVersion=x.x.x`.
// If it is not set, fall back to `{branch}-{hash}`.
fun Project.artifactVersion(project: Project): String =
    project.findProperty("artifactVersion")?.toString() ?: getBuildVersion().toString()

private fun String.runCommand(workingDir: java.io.File): String {
    try {
        val parts = this.split("\\s".toRegex())
        val proc = ProcessBuilder(*parts.toTypedArray())
            .directory(workingDir)
            .redirectOutput(ProcessBuilder.Redirect.PIPE)
            .redirectError(ProcessBuilder.Redirect.PIPE)
            .start()

        proc.waitFor(5, java.util.concurrent.TimeUnit.SECONDS)

        val errorStream = proc.errorStream.bufferedReader().readText()
        if (errorStream != "") throw RuntimeException(errorStream)

        val inputStream = proc.inputStream.bufferedReader().readText().trim()
        return inputStream
    } catch(e: java.io.IOException) {
        throw e
    }
}

private val VERSION_REGEX = "\\d+\\.\\d+\\.\\d+".toRegex()
private fun getReleaseVersion(): String = "git describe --always".runCommand(java.io.File("./")).replaceFirst("v", "")
private fun getBranch(): String = "git rev-parse --abbrev-ref HEAD".runCommand(java.io.File("./")).replace('/', '_')
private fun getBranchCommitHash(): String = "git log -1 --format=%h".runCommand(java.io.File("./"))

private fun getBuildVersion(): String {
    var version = getReleaseVersion()
    if (!VERSION_REGEX.matches(version)) {
        val branch = getBranch()
        val commit = getBranchCommitHash()
        version = "$branch-${commit}"
    }
    return version
}
