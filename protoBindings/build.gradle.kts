import java.nio.file.Paths

println("""
Welcome to Gradle ${gradle.gradleVersion} - http://www.gradle.org
Gradle home is set to: ${gradle.gradleHomeDir}
Gradle user directory is set to: ${gradle.gradleUserHomeDir}

Base directory: $projectDir
Running script ${relativePath(buildFile)}
""")

plugins {
    kotlin("jvm") version PluginVersions.Kotlin
    java
    id(PluginIds.Idea)
    id(PluginIds.TaskTree) version PluginVersions.TaskTree
    id(PluginIds.TestLogger) version PluginVersions.TestLogger apply false
    id(PluginIds.DependencyAnalysis) version PluginVersions.DependencyAnalysis
    id(PluginIds.Protobuf) version PluginVersions.Protobuf
    id(PluginIds.ProtobufRustGrpc)
    id(PluginIds.MavenPublish)
    id(PluginIds.Signing)
    id(PluginIds.NexusPublish) version PluginVersions.NexusPublish
}

allprojects {
    group = rootProject.name
    version = artifactVersion(this)

    repositories {
        mavenCentral()
        jcenter()
    }
}

configurations.all {
    resolutionStrategy {
        cacheDynamicVersionsFor(0, "seconds")
        cacheChangingModulesFor(0, "seconds")
    }
}

// Publishing
nexusPublishing {
    repositories {
        sonatype {
            nexusUrl.set(uri("https://s01.oss.sonatype.org/service/local/"))
            snapshotRepositoryUrl.set(uri("https://s01.oss.sonatype.org/content/repositories/snapshots/"))
            username.set(findProject("ossrhUsername")?.toString() ?: System.getenv("OSSRH_USERNAME"))
            password.set(findProject("ossrhPassword")?.toString() ?: System.getenv("OSSRH_PASSWORD"))
            stagingProfileId.set("3180ca260b82a7") // prevents querying for the staging profile id, performance optimization
        }
    }
}

subprojects {
    apply {
        plugin("java")
        plugin(PluginIds.Kotlin)
        plugin(PluginIds.Idea)
        plugin(PluginIds.TestLogger)
        plugin(PluginIds.Protobuf)
        plugin(PluginIds.MavenPublish)
        plugin(PluginIds.Signing)
    }

    repositories {
        mavenLocal()
        mavenCentral()
        jcenter()
        maven { url = project.uri("https://maven.java.net/content/groups/public") }
        maven { url = project.uri("https://oss.sonatype.org/content/repositories/snapshots") }
    }

    dependencies {
        api(Libraries.ProtobufJavaUtil)
        api(Libraries.GrpcKotlinStub)
        api(Libraries.GrpcProtobuf)

        if (JavaVersion.current().isJava9Compatible) {
            // Workaround for @javax.annotation.Generated
            // see: https://github.com/grpc/grpc-java/issues/3633
            api("javax.annotation:javax.annotation-api:1.3.1")
        }

        implementation(Libraries.KotlinReflect)
        implementation(Libraries.KotlinStdlib)
        implementation(Libraries.ProtobufKotlin)
        implementation(Libraries.GrpcStub)
        implementation(Libraries.GrpcNetty)
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

    plugins.withType<com.adarshr.gradle.testlogger.TestLoggerPlugin> {
        configure<com.adarshr.gradle.testlogger.TestLoggerExtension> {
            theme = com.adarshr.gradle.testlogger.theme.ThemeType.STANDARD
            showCauses = true
            slowThreshold = 1000
            showSummary = true
        }
    }

    // Generate sources Jar and Javadocs
    java {
        withJavadocJar()
        withSourcesJar()
    }

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
                    name.set("Cosmos Proto Bindings")
                    description.set("Protobuf bindings for JVM languages")
                    url.set("https://cosmos.network")

                    licenses {
                        license {
                            name.set("The Apache License, Version 2.0")
                            url.set("http://www.apache.org/licenses/LICENSE-2.0.txt")
                        }
                    }

                    developers {
                        developer {
                            id.set("egaxhaj-figure")
                            name.set("Ergels Gaxhaj")
                            email.set("egaxhaj@figure.com")
                        }
                    }

                    scm {
                        connection.set("git@github.com:cosmos/cosmos-sdk.git")
                        developerConnection.set("git@github.com/cosmos/cosmos-sdk.git")
                        url.set("https://github.com/cosmos/cosmos-sdk")
                    }
                }
            }
        }

        signing {
            sign(publishing.publications["mavenJava"])
        }
    }

}
