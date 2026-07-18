package com.jphat.buzzpi.di

import android.content.Context
import com.jphat.buzzpi.data.bpp.BppClient
import com.jphat.buzzpi.data.bpp.HandshakeHandler
import com.jphat.buzzpi.data.discovery.MdnsDiscovery
import com.jphat.buzzpi.data.repository.DeviceRepositoryImpl
import com.jphat.buzzpi.data.repository.SessionRepositoryImpl
import com.jphat.buzzpi.domain.repository.DeviceRepository
import com.jphat.buzzpi.domain.repository.SessionRepository
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import okhttp3.OkHttpClient
import java.util.concurrent.TimeUnit
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object AppModule {

    @Provides
    @Singleton
    fun provideOkHttpClient(): OkHttpClient {
        return OkHttpClient.Builder()
            .connectTimeout(10, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .writeTimeout(30, TimeUnit.SECONDS)
            .pingInterval(15, TimeUnit.SECONDS)
            .build()
    }

    @Provides
    @Singleton
    fun provideApplicationScope(): CoroutineScope {
        return CoroutineScope(SupervisorJob() + Dispatchers.Default)
    }

    @Provides
    @Singleton
    fun provideMdnsDiscovery(
        @ApplicationContext context: Context
    ): MdnsDiscovery {
        return MdnsDiscovery(context)
    }

    @Provides
    @Singleton
    fun provideBppClient(
        okHttpClient: OkHttpClient
    ): BppClient {
        return BppClient(okHttpClient)
    }

    @Provides
    @Singleton
    fun provideHandshakeHandler(
        bppClient: BppClient
    ): HandshakeHandler {
        return HandshakeHandler(bppClient)
    }

    @Provides
    @Singleton
    fun provideDeviceRepository(
        @ApplicationContext context: Context,
        mdnsDiscovery: MdnsDiscovery,
        bppClient: BppClient,
        handshakeHandler: HandshakeHandler,
        sessionRepositoryImpl: SessionRepositoryImpl
    ): DeviceRepository {
        return DeviceRepositoryImpl(
            mdnsDiscovery = mdnsDiscovery,
            bppClient = bppClient,
            handshakeHandler = handshakeHandler,
            sessionRepository = sessionRepositoryImpl,
            context = context
        )
    }

    @Provides
    @Singleton
    fun provideSessionRepository(
        @ApplicationContext context: Context
    ): SessionRepository {
        return SessionRepositoryImpl(context)
    }
}
