package com.jphat.buzzpi.ui.screens.filemanager

import android.app.Application
import android.content.Intent
import android.net.Uri
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.jphat.buzzpi.data.bpp.BppClient
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONArray
import org.json.JSONObject
import javax.inject.Inject

data class FileItem(
    val name: String,
    val path: String,
    val isDirectory: Boolean,
    val size: Long = 0,
    val modified: String = ""
)

data class FileManagerUiState(
    val currentPath: String = "/home",
    val files: List<FileItem> = emptyList(),
    val isLoading: Boolean = false,
    val isUploading: Boolean = false,
    val error: String? = null,
    val breadcrumb: List<String> = listOf("home")
)

@HiltViewModel
class FileManagerViewModel
@Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val bppClient: BppClient,
    private val application: Application
) : ViewModel() {

    private val deviceId: String = savedStateHandle["deviceId"] ?: ""
    private val _uiState = MutableStateFlow(FileManagerUiState())
    val uiState: StateFlow<FileManagerUiState> = _uiState.asStateFlow()

    init {
        browseDirectory("/home")
    }

    fun browseDirectory(path: String) {
        _uiState.value = _uiState.value.copy(
            isLoading = true,
            currentPath = path,
            error = null
        )

        viewModelScope.launch {
            try {
                val params = JSONObject().apply {
                    put("path", path)
                }
                val response = bppClient.sendRequest("file.browse", params)
                val resultJson = response.result?.let { JSONObject(it.decodeToString()) }
                val filesArray = resultJson?.optJSONArray("files") ?: JSONArray()
                val files = parseFiles(filesArray)
                val breadcrumb = buildBreadcrumb(path)

                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    files = files,
                    breadcrumb = breadcrumb
                )
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = e.message ?: "Failed to browse directory"
                )
            }
        }
    }

    fun navigateToFolder(folder: FileItem) {
        if (folder.isDirectory) {
            browseDirectory(folder.path)
        }
    }

    fun navigateUp() {
        val parent = _uiState.value.currentPath.substringBeforeLast("/", "")
        if (parent.isNotEmpty()) {
            browseDirectory(parent)
        }
    }

    fun navigateToBreadcrumb(index: Int) {
        val path = "/" + _uiState.value.breadcrumb.take(index + 1).joinToString("/")
        browseDirectory(path)
    }

    fun downloadFile(file: FileItem) {
        viewModelScope.launch {
            try {
                val params = JSONObject().apply {
                    put("path", file.path)
                }
                val response = bppClient.sendRequest("file.download", params)
                val resultJson = response.result?.let { JSONObject(it.decodeToString()) }
                val data = resultJson?.optString("data", "") ?: ""
                val mimeType = resultJson?.optString("mime_type", "application/octet-stream") ?: "application/octet-stream"

                if (data.isNotEmpty()) {
                    val intent = Intent(Intent.ACTION_VIEW).apply {
                        setDataAndType(Uri.parse(data), mimeType)
                        addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                    }
                    application.startActivity(intent)
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to download file"
                )
            }
        }
    }

    fun deleteFile(file: FileItem) {
        viewModelScope.launch {
            try {
                val params = JSONObject().apply {
                    put("path", file.path)
                }
                bppClient.sendRequest("file.delete", params)
                browseDirectory(_uiState.value.currentPath)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to delete file"
                )
            }
        }
    }

    fun createFolder(name: String) {
        viewModelScope.launch {
            try {
                val path = "${_uiState.value.currentPath}/$name"
                val params = JSONObject().apply {
                    put("path", path)
                }
                bppClient.sendRequest("file.mkdir", params)
                browseDirectory(_uiState.value.currentPath)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to create folder"
                )
            }
        }
    }

    fun renameFile(file: FileItem, newName: String) {
        viewModelScope.launch {
            try {
                val params = JSONObject().apply {
                    put("path", file.path)
                    put("new_name", newName)
                }
                bppClient.sendRequest("file.rename", params)
                browseDirectory(_uiState.value.currentPath)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    error = e.message ?: "Failed to rename file"
                )
            }
        }
    }

    fun uploadFile(uri: Uri) {
        _uiState.value = _uiState.value.copy(isUploading = true)
        viewModelScope.launch {
            try {
                val inputStream = application.contentResolver.openInputStream(uri)
                val bytes = inputStream?.readBytes() ?: throw Exception("Failed to read file")
                inputStream.close()

                val fileName = uri.lastPathSegment ?: "upload"
                val path = "${_uiState.value.currentPath}/$fileName"

                val params = JSONObject().apply {
                    put("path", path)
                    put("data", android.util.Base64.encodeToString(bytes, android.util.Base64.DEFAULT))
                }
                bppClient.sendRequest("file.upload", params)
                _uiState.value = _uiState.value.copy(isUploading = false)
                browseDirectory(_uiState.value.currentPath)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isUploading = false,
                    error = e.message ?: "Failed to upload file"
                )
            }
        }
    }

    private fun parseFiles(array: JSONArray): List<FileItem> {
        val files = mutableListOf<FileItem>()
        for (i in 0 until array.length()) {
            val obj = array.getJSONObject(i)
            files.add(FileItem(
                name = obj.getString("name"),
                path = obj.getString("path"),
                isDirectory = obj.getBoolean("is_directory"),
                size = obj.optLong("size", 0),
                modified = obj.optString("modified", "")
            ))
        }
        return files.sortedWith(compareByDescending<FileItem> { it.isDirectory }.thenBy { it.name })
    }

    private fun buildBreadcrumb(path: String): List<String> {
        if (path == "/") return listOf("/")
        return path.split("/").filter { it.isNotEmpty() }
    }

    fun clearError() {
        _uiState.value = _uiState.value.copy(error = null)
    }
}
