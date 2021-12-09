package rust

import org.gradle.api.Plugin
import org.gradle.api.Project
import org.gradle.kotlin.dsl.register

/**
 * Custom gradle plugin to download Provenance, Cosmos, and CosmWasm/wasmd protobuf files.
 *
 */
class ProtobufRustGrpcPlugin : Plugin<Project> {
    override fun apply(project: Project) {
        project.tasks.register(
            "protobuf-rust-grpc",
            ProtobufRustGrpcTask::class
        ) {
            this.group = "protobuf"
            this.description = "Generate proto source files for Rust."
        }
    }
}
