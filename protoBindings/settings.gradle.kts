// The name of your project
rootProject.name = "io.provenance"
// By default, the artifact version number is derived
// from the latest SDK release. Uncomment the line below
// if you want to have a different release path for the
// proto artifacts.
//artifactVersion = "1.0-SNAPSHOT"

// Dynamically include subprojects in `bindings/`.
// Uses the folder name as the project name.
File(rootDir, "bindings").walk().filter {
    it.isDirectory && File(it, "build.gradle.kts").isFile
}.forEach {
    // ignore rust bindings for now until it is ready.
    if (it.name != "rust") {
        include(it.name)
        project(":${it.name}").projectDir = it
    }
}
