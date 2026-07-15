# BuzzPi ProGuard Rules
# This file is intentionally empty. Add project-specific rules as needed.

# Keep Hilt generated classes
-keepclassmembers class * {
    @dagger.hilt.android.EarlyEntryPoint <fields>;
    @dagger.hilt.android.EarlyEntryPoint <methods>;
}

# Keep Kotlin serialization
-keepattributes *Annotation*, InnerClasses
-dontnote kotlinx.serialization.AnnotationsKt

# Keep WebSocket/OkHttp
-dontwarn okhttp3.**
-dontwarn okio.**
