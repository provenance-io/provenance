import com.google.protobuf.gradle.generateProtoTasks
import com.google.protobuf.gradle.id
import com.google.protobuf.gradle.plugins
import com.google.protobuf.gradle.protobuf
import com.google.protobuf.gradle.protoc
import org.jetbrains.kotlin.gradle.tasks.KotlinCompile
import java.nio.file.Paths

plugins {
    // Apply the java-library plugin for API and implementation separation.
    `java-library`
    kotlin("jvm") version PluginVersions.Kotlin
    id(PluginIds.Protobuf) version PluginVersions.Protobuf
    id(PluginIds.MavenPublish)
    id(PluginIds.Signing)
    id(PluginIds.KtLint) version PluginVersions.KtLint
}

group = project.property("group.id") as String
version = artifactVersion(rootProject)

repositories {
    // Use Maven Central for resolving dependencies.
    mavenCentral()
}

dependencies {
    // Use JUnit test framework.
    testImplementation(Libraries.JUnit)

    // This dependency is exported to consumers, that is to say found on their compile classpath.
    api(Libraries.ProtobufJavaUtil)
    api(Libraries.GrpcKotlinStub)
    api(Libraries.GrpcProtobuf)
    if (JavaVersion.current().isJava9Compatible) {
        // Workaround for @javax.annotation.Generated
        // see: https://github.com/grpc/grpc-java/issues/3633
        api("javax.annotation:javax.annotation-api:1.3.1")
    }

    // This dependency is used internally, and not exposed to consumers on their own compile classpath.
    implementation(Libraries.KotlinReflect)
    implementation(Libraries.KotlinStdlib)
    implementation(Libraries.ProtobufKotlin)
    implementation(Libraries.GrpcStub)
}

tasks.jar {
    archiveBaseName.set("proto-${project.name}")
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

// Protobuf file source directories
sourceSets.main {
    val protoDirs = (project.property("protoDirs") as String).split(",")
        .map {
            var path = it.trim()
            if (File(it).isAbsolute()) {
                path
            } else {
                path = Paths.get(rootProject.projectDir.toString(), path).toString()
                // Normalize relative paths. Example: foo/../bar/baz => bar/baz
                File(path).normalize()
            }
        }
    proto.srcDirs(protoDirs)
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

// Generate sources Jar and Javadocs
java {
    withJavadocJar()
    withSourcesJar()
}

// Maven publishing
publishing {
    publications {
        create<MavenPublication>("mavenJava") {
            from(components["java"])

            afterEvaluate {
                groupId = project.group.toString()
                artifactId = tasks.jar.get().archiveBaseName.get()
                version = tasks.jar.get().archiveVersion.get()
            }

            pom {
                name.set(project.property("pom.name") as String)
                description.set(project.property("pom.description") as String)
                url.set(project.property("pom.url") as String)

                licenses {
                    license {
                        name.set(project.property("license.name") as String)
                        url.set(project.property("license.url") as String)
                    }
                }

                developers {
                    developer {
                        id.set(project.property("developer.id") as String)
                        name.set(project.property("developer.name") as String)
                        email.set(project.property("developer.email") as String)
                    }
                }

                scm {
                    connection.set(project.property("scm.connection") as String)
                    developerConnection.set(project.property("scm.developerConnection") as String)
                    url.set(project.property("scm.url") as String)
                }
            }
        }
    }

    signing {
        sign(publishing.publications["mavenJava"])
    }
}
