object PluginIds { // please keep this sorted in sections
    // Kotlin
    const val Kotlin = "kotlin"

    const val Protobuf = "com.google.protobuf"
    const val Grpc = "grpc"
    const val GrpcKt = "grpckt"

    // Publishing
    const val MavenPublish = "maven-publish"
    const val Signing = "signing"
    const val NexusPublish = "io.github.gradle-nexus.publish-plugin"

    // Linting (Kotlin)
    const val Spotless = "com.diffplug.spotless"
}

object PluginVersions { // please keep this sorted in sections
    // Kotlin
    const val Kotlin = "1.9.25"

    // Protobuf
    const val Protobuf = "0.9.5"

    // Publishing
    const val NexusPublish = "1.1.0"

    // KtLint
    const val Spotless = "7.1.0"
}

object Versions {
    // kotlin
    const val Kotlin = PluginVersions.Kotlin

    // Protobuf & gRPC
    // Use .5 for compatibility with older protobuf runtimes (see https://protobuf.dev/news/2025-01-23/)
    const val Protobuf = "3.25.5"
    const val Grpc = "1.73.0"
    const val KotlinGrpc = "1.4.3"

    // Testing
    const val JUnit = "4.13.2"
}

object Libraries {
    // Kotlin
    const val KotlinReflect = "org.jetbrains.kotlin:kotlin-reflect:${Versions.Kotlin}"
    const val KotlinStdlib = "org.jetbrains.kotlin:kotlin-stdlib:${Versions.Kotlin}"

    // Protobuf
    const val ProtobufJavaUtil = "com.google.protobuf:protobuf-java-util:${Versions.Protobuf}"
    const val ProtobufKotlin = "com.google.protobuf:protobuf-kotlin:${Versions.Protobuf}"
    const val GrpcProtobuf = "io.grpc:grpc-protobuf:${Versions.Grpc}"
    const val GrpcStub = "io.grpc:grpc-stub:${Versions.Grpc}"
    const val GrpcKotlinStub = "io.grpc:grpc-kotlin-stub:${Versions.KotlinGrpc}"
    const val ProtocArtifact = "com.google.protobuf:protoc:${Versions.Protobuf}"
    const val GrpcArtifact = "io.grpc:protoc-gen-grpc-java:${Versions.Grpc}"
    const val GrpcKotlinArtifact = "io.grpc:protoc-gen-grpc-kotlin:${Versions.KotlinGrpc}:jdk8@jar"

    // Testing
    const val JUnit = "junit:junit:${Versions.JUnit}"
}
