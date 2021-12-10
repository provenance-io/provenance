
// Register task
tasks.register<rust.ProtobufRustGrpcTask>("generateRustGrpc")

// Remove Maven publishing
tasks.withType<PublishToMavenRepository>().configureEach { setProperty("enabled", false) }