import com.codingfeline.buildkonfig.compiler.FieldSpec
import org.jetbrains.kotlin.gradle.ExperimentalWasmDsl
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    alias(libs.plugins.kotlinMultiplatform)
    alias(libs.plugins.androidMultiplatformLibrary)
    alias(libs.plugins.composeMultiplatform)
    alias(libs.plugins.composeCompiler)
    alias(libs.plugins.apollo)
    alias(libs.plugins.buildkonfig)
}

val envFile = rootProject.file("../.env")

val env = mutableMapOf<String, String>()

if (envFile.exists()) {
    envFile.forEachLine { line ->
        val trimmed = line.trim()

        if (
            trimmed.isNotEmpty() &&
            !trimmed.startsWith("#") &&
            trimmed.contains("=")
        ) {
            val parts = trimmed.split("=", limit = 2)

            env[parts[0].trim()] = parts[1].trim()
        }
    }
}

val apiBaseUrl = env["API_BASE_URL"] ?: "" // Fallback to empty string if missing to avoid sync error

buildkonfig {
    packageName = "com.example.kalasetu"

    defaultConfigs {
        buildConfigField(
            FieldSpec.Type.STRING,
            "API_BASE_URL",
            apiBaseUrl
        )
    }
}

kotlin {
    listOf(
        iosArm64(),
        iosSimulatorArm64(),
    ).forEach { iosTarget ->
        iosTarget.binaries.framework {
            baseName = "Shared"
            isStatic = true
        }
    }

    jvm()

    js {
        browser()
    }

    @OptIn(ExperimentalWasmDsl::class)
    wasmJs {
        browser()
    }

    android {
        namespace = "com.example.kalasetu.app.shared"
        compileSdk = libs.versions.android.compileSdk.get().toInt()
        minSdk = libs.versions.android.minSdk.get().toInt()

        compilerOptions {
            jvmTarget = JvmTarget.JVM_11
        }

        androidResources {
            enable = true
        }

        withHostTest {
            isIncludeAndroidResources = true
        }
    }

    sourceSets {
        androidMain.dependencies {
            implementation(libs.compose.uiToolingPreview)
            implementation(libs.peekaboo.image.picker)
            implementation(libs.peekaboo.ui)
            implementation(libs.androidx.activity.compose)
            implementation(libs.ktor.client.okhttp)
        }

        commonMain.dependencies {
            api(projects.core)
            implementation(libs.kotlinx.coroutines.core)
            implementation(libs.compose.runtime)
            implementation(libs.compose.foundation)
            implementation(libs.compose.material3)
            implementation(libs.compose.ui)
            implementation(libs.compose.components.resources)
            implementation(libs.compose.uiToolingPreview)
            implementation(libs.compose.materialIconsCore)
            implementation(libs.compose.materialIconsExtended)
            implementation("org.jetbrains.compose.ui:ui-backhandler:1.11.1")
            implementation(libs.androidx.lifecycle.viewmodelCompose)
            implementation(libs.androidx.lifecycle.runtimeCompose)
            implementation(libs.coil.compose)
            implementation(libs.kotlinx.datetime)
            implementation(libs.apollo.runtime)
            implementation(libs.filekit.core)
            implementation(libs.filekit.compose)
            implementation(libs.ktor.client.core)
            implementation(libs.ktor.client.content.negotiation)
            implementation(libs.ktor.serialization.kotlinx.json)
            implementation(libs.coil.network.ktor3)
        }

        iosMain.dependencies {
            implementation(libs.peekaboo.image.picker)
            implementation(libs.peekaboo.ui)
        }

        commonTest.dependencies {
            implementation(libs.kotlin.test)
            implementation(libs.kotlinx.coroutines.test)
        }

        jsMain.dependencies {
            implementation(libs.wrappers.browser)
        }
    }
}

dependencies {
    androidRuntimeClasspath(libs.compose.uiTooling)
}

apollo {
    service("service") {
        packageName.set("com.example.kalasetu")
        mapScalar("Upload", "com.apollographql.apollo.api.Upload")
    }
}