plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
}

android {
    namespace = "com.takahashi.yutecoin"
    compileSdk {
        version = release(36) {
            minorApiLevel = 1
        }
    }

    defaultConfig {
        applicationId = "com.takahashi.yutecoin"
        minSdk = 24
        targetSdk = 36
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        // Base URL for API:
        //   Emulator: "http://10.0.2.2:3010/"
        //   Physical device + adb reverse: "http://127.0.0.1:3010/"
        //   Physical device + LAN: "http://<LAN_IP>:3010/"
        buildConfigField("String", "API_BASE_URL", "\"http://127.0.0.1:3010/\"")
    }

    buildTypes {
        release {
            optimization {
                enable = false
            }
        }
    }
    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_11
        targetCompatibility = JavaVersion.VERSION_11
    }
    buildFeatures {
        compose = true
        buildConfig = true
    }
    testOptions {
        unitTests.isReturnDefaultValues = true
    }
    packaging {
        resources {
            excludes += setOf(
                "META-INF/DISCLAIMER",
                "META-INF/DEPENDENCIES"
            )
        }
    }
}

dependencies {
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.graphics)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)

    // Networking
    implementation(libs.retrofit)
    implementation(libs.retrofit.converter.gson)
    implementation(libs.okhttp)
    implementation(libs.okhttp.logging)
    implementation(libs.gson)

    // Crypto (BIP39, BIP44, ECDSA, Keccak256 via BouncyCastle)
    implementation(libs.bouncycastle)

    // Navigation
    implementation(libs.navigation.compose)
    implementation(libs.lifecycle.viewmodel.compose)

    // DI
    implementation(libs.koin.android)
    implementation(libs.koin.compose)

    // Persistence
    implementation(libs.datastore)
    implementation(libs.security.crypto)

    // Coroutines
    implementation(libs.coroutines)

    // Test
    testImplementation(libs.junit)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.compose.ui.test.junit4)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(libs.androidx.junit)
    debugImplementation(libs.androidx.compose.ui.test.manifest)
    debugImplementation(libs.androidx.compose.ui.tooling)
}
