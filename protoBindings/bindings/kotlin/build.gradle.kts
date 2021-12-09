import com.google.protobuf.gradle.generateProtoTasks
import com.google.protobuf.gradle.id
import com.google.protobuf.gradle.ofSourceSet
import com.google.protobuf.gradle.plugins
import com.google.protobuf.gradle.protobuf
import com.google.protobuf.gradle.protoc
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile

tasks.jar {
    baseName = "${rootProject.name}-proto-kotlin"
}

tasks.withType<Javadoc> { enabled = true }

tasks.withType<JavaCompile> {
    sourceCompatibility = JavaVersion.VERSION_11.toString()
    targetCompatibility = sourceCompatibility
}

tasks.withType<KotlinCompile> {
    kotlinOptions {
        freeCompilerArgs = listOf("-Xjsr305=strict", "-Xopt-in=kotlin.RequiresOptIn")
        jvmTarget = "11"
        languageVersion = "1.5"
        apiVersion = "1.5"
    }
}

// For more advanced options see: https://github.com/google/protobuf-gradle-plugin
protobuf {
    protoc {
        // The artifact spec for the Protobuf Compiler
        artifact = Libraries.ProtocArtifact
    }
    plugins {
        // Optional: an artifact spec for a protoc plugin, with "grpc" as
        // the identifier, which can be referred to in the "plugins"
        // container of the "generateProtoTasks" closure.
        id(PluginIds.Grpc) {
            artifact = Libraries.GrpcArtifact
        }
        id(PluginIds.GrpcKt) {
            artifact = Libraries.GrpcKotlinArtifact
        }
    }
    generateProtoTasks {
        all().forEach { task ->
            task.plugins {
                id(PluginIds.Grpc)
                id(PluginIds.GrpcKt)
            }
            task.builtins {
                id(PluginIds.Kotlin)
            }

            task.generateDescriptorSet = true
        }
    }
}
