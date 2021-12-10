object PluginIds { // please keep this sorted in sections
    // Kotlin
    const val Kotlin = "kotlin"
    const val KotlinSpring = "plugin.spring"
    const val Kapt = "kapt"

    // 3rd Party
    const val Flyway = "org.flywaydb.flyway"
    const val Idea = "idea"
    const val TaskTree = "com.dorongold.task-tree"
    const val TestLogger = "com.adarshr.test-logger"
    const val DependencyAnalysis = "com.autonomousapps.dependency-analysis"
    const val GoryLenkoGitProps = "com.gorylenko.gradle-git-properties"

    const val SpringDependency = "io.spring.dependency-management"
    const val SpringBoot = "org.springframework.boot"
    const val Protobuf = "com.google.protobuf"
    const val Grpc = "grpc"
    const val GrpcKt = "grpckt"

    // Rust
    const val ProtobufRustGrpc = "rust.protobuf-rust-grpc"

    // Publishing
    const val MavenPublish = "maven-publish"
    const val Signing = "signing"
    const val NexusPublish = "io.github.gradle-nexus.publish-plugin"
}

object PluginVersions { // please keep this sorted in sections
    // Kotlin
    const val Kotlin = "1.5.30"

    // 3rd Party
    const val TaskTree = "2.1.0"
    const val TestLogger = "2.1.1"
    const val DependencyAnalysis = "0.56.0"
    const val GoryLenkoGitProps = "1.5.2"

    const val Protobuf = "0.8.18"

    // Publishing
    const val NexusPublish = "1.1.0"
}

object Versions {
    // kotlin
    const val Kotlin = PluginVersions.Kotlin
    const val KotlinXCoroutines = "1.5.2"

    // 3rd Party
    const val ApacheCommonsText = "1.9"
    const val KaseChange = "1.3.0"
    const val Protobuf = "3.19.1"
    const val Grpc = "1.40.1"
    const val KotlinGrpc = "1.2.0"
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
    const val GrpcNetty = "io.grpc:grpc-netty:${Versions.Grpc}"

}

// gradle configurations
const val kapt = "kapt"
const val api = "api"
