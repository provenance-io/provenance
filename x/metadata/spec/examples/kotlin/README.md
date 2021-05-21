# Metadata Kotlin Examples

This README is only here for developer troubleshooting information.
Individual code examples should be discussed in the main spec files.

## IntelliJ Troubleshooting

Problem:

Either the .kt files have a lot of red or IntelliJ doesn't give an option to run the unit tests:

Solution:

1. Open the `x/metadata/spec/examples/kotlin/build.gradle.kts` file.
1. There should be a banner at the top indicating an issue with code insight.
1. Click the "Link Gradle Project" link on the right end of that banner.
1. Navigate to and select that same `x/metadata/spec/examples/kotlin/build.gradle.kts` file.
1. Wait for the loading/indexing to finish.

