package io.provenance

import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.kotlin.dsl.register

/**
 * Custom gradle plugin to download Provenance, Cosmos, and CosmWasm/wasmd protobuf files.
 *
 */
class DownloadProtosPlugin : Plugin<Project> {
    override fun apply(project: Project) {
        project.tasks.register(
            "downloadProtos",
            DownloadProtosTask::class
        ) {
            this.group = "protobuf"
            this.description =
                """Downloads Cosmos and CosmWasm protobuf files. 
                   Specify Cosmos, and CosmWasm versions: 
                       --cosmos-version vX.Y.Z --wasmd-version vX.Y.Z
                   Version information can be found at: 
                       https://github.com/cosmos/cosmos-sdk/releases, and 
                       https://github.com/CosmWasm/wasmd/tags."""
        }
    }
}
