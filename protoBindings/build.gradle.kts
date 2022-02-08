
plugins {
    /*
     * Nexus publishing plugin cannot exist in sub projects.
     * See https://github.com/gradle-nexus/publish-plugin/issues/81
     */
    id(PluginIds.NexusPublish) version PluginVersions.NexusPublish
}

// Publishing
nexusPublishing {
    repositories {
        sonatype {
            nexusUrl.set(uri(project.property("nexus.url") as String))
            snapshotRepositoryUrl.set(uri(project.property("nexus.snapshot.repository.url") as String))
            username.set(findProject(project.property("nexus.username") as String)?.toString() ?: System.getenv("OSSRH_USERNAME"))
            password.set(findProject(project.property("nexus.password") as String)?.toString() ?: System.getenv("OSSRH_PASSWORD"))
            stagingProfileId.set(project.property("nexus.staging.profile.id") as String) // prevents querying for the staging profile id, performance optimization
        }
    }
}
