// https://docs.gradle.org/current/userguide/kotlin_dsl.html#sec:kotlin-dsl_plugin
plugins {
  `kotlin-dsl`
}

repositories {
    mavenCentral()
}

dependencies {
    api("org.apache.commons:commons-compress:1.20")
    api("commons-io:commons-io:2.6")
    api("org.apache.httpcomponents:httpclient:4.5.12")
}
