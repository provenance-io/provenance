package rust

import org.gradle.api.DefaultTask
import org.gradle.api.tasks.TaskAction

/**
 * Custom gradle task to generate Rust protobuf source files.
 */
open class ProtobufRustGrpcTask : DefaultTask() {
    /**
     * Generate proto source files for Rust
     *
     */
    @TaskAction
    fun generateRustGrpc() {
        throw NotImplementedError(
            message = "*** :generateRustGrpc is not yet supported! ***")
    }
}
