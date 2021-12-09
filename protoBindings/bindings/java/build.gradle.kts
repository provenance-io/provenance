import com.google.protobuf.gradle.generateProtoTasks
import com.google.protobuf.gradle.id
import com.google.protobuf.gradle.ofSourceSet
import com.google.protobuf.gradle.plugins
import com.google.protobuf.gradle.protobuf
import com.google.protobuf.gradle.protoc

tasks.jar {
    baseName = "proto-${project.name}"
}

tasks.withType<Javadoc> { enabled = true }

tasks.withType<JavaCompile> {
    sourceCompatibility = JavaVersion.VERSION_11.toString()
    targetCompatibility = sourceCompatibility
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
    }
    generateProtoTasks {
        all().forEach { task ->
            task.plugins {
                id(PluginIds.Grpc)
            }
        }
    }
}
