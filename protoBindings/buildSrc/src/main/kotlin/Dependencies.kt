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

    // User defined plugins in `buildSrc/src/main/kotlin/`
    const val ProtobufRustGrpc = "rust.protobuf-rust-grpc"

    // Linting (Kotlin)
    const val KtLint = "org.jlleitschuh.gradle.ktlint"
}

object PluginVersions { // please keep this sorted in sections
    // Kotlin
    const val Kotlin = "1.5.30"

    // Protobuf
    const val Protobuf = "0.8.18"

    // Publishing
    const val NexusPublish = "1.1.0"

    // KtLint
    const val KtLint = "10.2.0"
}

object Versions {
    // kotlin
    const val Kotlin = PluginVersions.Kotlin
    const val KotlinXCoroutines = "1.5.2"

    // Protobuf & gRPC
    const val Protobuf = "3.19.1"
    const val Grpc = "1.40.1"
    const val KotlinGrpc = "1.2.0"

    // Testing
    const val JUnit = "4.13.2"
}

object Libraries {
    // Kotlin
    const val KotlinReflect = "org.jetbrains.kotlin:kotlin-reflect:${Versions.Kotlin}"
    const val KotlinStdlib = "org.jetbrains.kotlin:kotlin-stdlib:${Versions.Kotlin}"
    const val KotlinXCoRoutinesCore = "org.jetbrains.kotlinx:kotlinx-coroutines-core:${Versions.KotlinXCoroutines}"
    const val KotlinXCoRoutinesGuava = "org.jetbrains.kotlinx:kotlinx-coroutines-guava:${Versions.KotlinXCoroutines}"

    // Protobuf
    const val ProtobufJavaUtil = "com.google.protobuf:protobuf-java-util:${Versions.Protobuf}"
    const val ProtobufKotlin = "com.google.protobuf:protobuf-kotlin:${Versions.Protobuf}"
    const val GrpcProtobuf = "io.grpc:grpc-protobuf:${Versions.Grpc}"
    const val GrpcStub = "io.grpc:grpc-stub:${Versions.Grpc}"
    const val GrpcKotlinStub = "io.grpc:grpc-kotlin-stub:${Versions.KotlinGrpc}"
    const val ProtocArtifact = "com.google.protobuf:protoc:${Versions.Protobuf}"
    const val GrpcArtifact = "io.grpc:protoc-gen-grpc-java:${Versions.Grpc}"
    const val GrpcKotlinArtifact = "io.grpc:protoc-gen-grpc-kotlin:${Versions.KotlinGrpc}:jdk7@jar"

    // Testing
    const val JUnit = "junit:junit:${Versions.JUnit}"
}
