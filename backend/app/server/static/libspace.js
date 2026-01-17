console.log("libspace.js/start");

const spaceRedirrectToAuth = (redirectBackUrl, actualPage) => {
    const prePageUrl = new URL('/zz/pages/auth/space/in_space/pre_page', window.location.origin);
    prePageUrl.searchParams.set('redirect_back_url', redirectBackUrl);
    prePageUrl.searchParams.set('actual_page', actualPage);
    window.location.href = prePageUrl.toString();
}

const spaceGetToken = (key) => {
    return localStorage.getItem(`${key}_space_token`);
}








/*
Usage:
    <div className="p-4">
        <button className="w-full bg-blue-600 text-white px-4 py-2 rounded-lg hover:bg-blue-700"
        onClick={() => {

            if (typeof window === 'undefined') return;

            if (!window.spaceFilePicker) return;
            if (!window.spaceGetToken) return;

            const token = window.spaceGetToken('cimple-xxx');
            if (!token) return;

            const picker = window.spaceFilePicker(token);
            if (!picker) return;
            picker.showModal((file) => {
                console.log(file);
            })

        }}
        
        >
            Show File Picker
        </button>
    </div>

*/

const spaceFilePicker = (spaceToken) => {
    let currentPath = '';
    let selectedFile = null;
    let onSelectCallback = null;
    let modal = null;
    let currentMode = 'browse'; // 'browse' or 'upload'
    let filesToUpload = [];

    const createModal = () => {
        // Create modal overlay
        const overlay = document.createElement('div');
        overlay.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 10000;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
        `;

        // Create modal content
        const content = document.createElement('div');
        content.style.cssText = `
            background: white;
            border-radius: 8px;
            width: 90%;
            max-width: 800px;
            max-height: 90vh;
            display: flex;
            flex-direction: column;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
        `;

        // Header
        const header = document.createElement('div');
        header.style.cssText = `
            padding: 20px;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            justify-content: space-between;
            align-items: center;
        `;
        const title = document.createElement('h2');
        title.textContent = 'File Manager';
        title.style.cssText = 'margin: 0; font-size: 20px; font-weight: 600;';
        const closeBtn = document.createElement('button');
        closeBtn.textContent = '×';
        closeBtn.style.cssText = `
            background: none;
            border: none;
            font-size: 28px;
            cursor: pointer;
            color: #666;
            padding: 0;
            width: 32px;
            height: 32px;
            line-height: 1;
        `;
        closeBtn.onclick = () => closeModal();
        header.appendChild(title);
        header.appendChild(closeBtn);

        // Tabs
        const tabs = document.createElement('div');
        tabs.style.cssText = `
            display: flex;
            border-bottom: 1px solid #e0e0e0;
            background: #f8f9fa;
        `;
        const browseTab = document.createElement('button');
        browseTab.id = 'space-picker-tab-browse';
        browseTab.textContent = 'Browse';
        browseTab.style.cssText = `
            flex: 1;
            padding: 12px 20px;
            border: none;
            background: transparent;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            color: #666;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        `;
        const uploadTab = document.createElement('button');
        uploadTab.id = 'space-picker-tab-upload';
        uploadTab.textContent = 'Upload';
        uploadTab.style.cssText = `
            flex: 1;
            padding: 12px 20px;
            border: none;
            background: transparent;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            color: #666;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        `;
        browseTab.onclick = () => switchMode('browse');
        uploadTab.onclick = () => switchMode('upload');
        tabs.appendChild(browseTab);
        tabs.appendChild(uploadTab);

        // Breadcrumb (for browse mode)
        const breadcrumb = document.createElement('div');
        breadcrumb.id = 'space-picker-breadcrumb';
        breadcrumb.style.cssText = `
            padding: 12px 20px;
            border-bottom: 1px solid #e0e0e0;
            display: flex;
            gap: 8px;
            align-items: center;
            flex-wrap: wrap;
            background: #f8f9fa;
        `;

        // File list container (for browse mode)
        const fileList = document.createElement('div');
        fileList.id = 'space-picker-list';
        fileList.style.cssText = `
            flex: 1;
            overflow-y: auto;
            padding: 8px;
            min-height: 300px;
        `;

        // Upload container (for upload mode)
        const uploadContainer = document.createElement('div');
        uploadContainer.id = 'space-picker-upload';
        uploadContainer.style.cssText = `
            flex: 1;
            overflow-y: auto;
            padding: 20px;
            min-height: 300px;
            display: none;
        `;

        // Upload area
        const uploadArea = document.createElement('div');
        uploadArea.style.cssText = `
            border: 2px dashed #ddd;
            border-radius: 8px;
            padding: 40px;
            text-align: center;
            cursor: pointer;
            transition: all 0.2s;
            background: #fafafa;
        `;
        uploadArea.onmouseover = () => {
            uploadArea.style.borderColor = '#007bff';
            uploadArea.style.background = '#f0f7ff';
        };
        uploadArea.onmouseout = () => {
            uploadArea.style.borderColor = '#ddd';
            uploadArea.style.background = '#fafafa';
        };

        const uploadIcon = document.createElement('div');
        uploadIcon.textContent = '📤';
        uploadIcon.style.cssText = 'font-size: 48px; margin-bottom: 12px;';
        const uploadText = document.createElement('div');
        uploadText.textContent = 'Click to select files or drag and drop';
        uploadText.style.cssText = 'color: #666; font-size: 14px; margin-bottom: 8px;';
        const uploadHint = document.createElement('div');
        uploadHint.textContent = 'You can select multiple files';
        uploadHint.style.cssText = 'color: #999; font-size: 12px;';

        const fileInput = document.createElement('input');
        fileInput.type = 'file';
        fileInput.multiple = true;
        fileInput.style.cssText = 'display: none;';
        fileInput.onchange = (e) => {
            if (e.target.files) {
                addFilesToUpload(Array.from(e.target.files));
            }
        };

        uploadArea.onclick = () => fileInput.click();
        uploadArea.appendChild(uploadIcon);
        uploadArea.appendChild(uploadText);
        uploadArea.appendChild(uploadHint);
        uploadArea.appendChild(fileInput);

        // Drag and drop
        uploadArea.ondragover = (e) => {
            e.preventDefault();
            uploadArea.style.borderColor = '#007bff';
            uploadArea.style.background = '#f0f7ff';
        };
        uploadArea.ondragleave = () => {
            uploadArea.style.borderColor = '#ddd';
            uploadArea.style.background = '#fafafa';
        };
        uploadArea.ondrop = (e) => {
            e.preventDefault();
            uploadArea.style.borderColor = '#ddd';
            uploadArea.style.background = '#fafafa';
            if (e.dataTransfer.files) {
                addFilesToUpload(Array.from(e.dataTransfer.files));
            }
        };

        // Files to upload list
        const filesList = document.createElement('div');
        filesList.id = 'space-picker-upload-files';
        filesList.style.cssText = 'margin-top: 20px;';

        uploadContainer.appendChild(uploadArea);
        uploadContainer.appendChild(filesList);

        // Footer with actions
        const footer = document.createElement('div');
        footer.style.cssText = `
            padding: 16px 20px;
            border-top: 1px solid #e0e0e0;
            display: flex;
            justify-content: flex-end;
            gap: 12px;
        `;
        const cancelBtn = document.createElement('button');
        cancelBtn.textContent = 'Cancel';
        cancelBtn.style.cssText = `
            padding: 8px 16px;
            border: 1px solid #ddd;
            background: white;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        `;
        cancelBtn.onclick = () => closeModal();
        const selectBtn = document.createElement('button');
        selectBtn.textContent = 'Select';
        selectBtn.id = 'space-picker-select-btn';
        selectBtn.style.cssText = `
            padding: 8px 16px;
            border: none;
            background: #007bff;
            color: white;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
        `;
        selectBtn.onclick = () => {
            if (selectedFile) {
                if (onSelectCallback) {
                    onSelectCallback(selectedFile);
                }
                closeModal();
            }
        };
        selectBtn.disabled = true;

        const uploadBtn = document.createElement('button');
        uploadBtn.textContent = 'Upload';
        uploadBtn.id = 'space-picker-upload-btn';
        uploadBtn.style.cssText = `
            padding: 8px 16px;
            border: none;
            background: #28a745;
            color: white;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            display: none;
        `;
        uploadBtn.onclick = () => uploadFiles();
        uploadBtn.disabled = true;

        footer.appendChild(cancelBtn);
        footer.appendChild(selectBtn);
        footer.appendChild(uploadBtn);

        content.appendChild(header);
        content.appendChild(tabs);
        content.appendChild(breadcrumb);
        content.appendChild(fileList);
        content.appendChild(uploadContainer);
        content.appendChild(footer);
        overlay.appendChild(content);

        // Close on overlay click
        overlay.onclick = (e) => {
            if (e.target === overlay) closeModal();
        };

        return overlay;
    };

    const updateBreadcrumb = () => {
        const breadcrumb = document.getElementById('space-picker-breadcrumb');
        if (!breadcrumb) return;

        breadcrumb.innerHTML = '';
        const parts = currentPath.split('/').filter(p => p);

        // Root link
        const rootLink = document.createElement('span');
        rootLink.textContent = 'Home';
        rootLink.style.cssText = `
            cursor: pointer;
            color: #007bff;
            padding: 4px 8px;
            border-radius: 4px;
        `;
        rootLink.onclick = () => navigateToPath('');
        rootLink.onmouseover = () => rootLink.style.background = '#e7f3ff';
        rootLink.onmouseout = () => rootLink.style.background = 'transparent';
        breadcrumb.appendChild(rootLink);

        // Path parts
        let pathSoFar = '';
        parts.forEach((part, index) => {
            const sep = document.createElement('span');
            sep.textContent = ' / ';
            sep.style.cssText = 'color: #999; margin: 0 4px;';
            breadcrumb.appendChild(sep);

            const link = document.createElement('span');
            link.textContent = part;
            link.style.cssText = `
                cursor: pointer;
                color: #007bff;
                padding: 4px 8px;
                border-radius: 4px;
            `;
            pathSoFar += '/' + part;
            link.onclick = () => navigateToPath(pathSoFar);
            link.onmouseover = () => link.style.background = '#e7f3ff';
            link.onmouseout = () => link.style.background = 'transparent';
            breadcrumb.appendChild(link);
        });
    };

    const formatFileSize = (bytes) => {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i];
    };

    const renderFileList = (files) => {
        const fileList = document.getElementById('space-picker-list');
        if (!fileList) return;

        fileList.innerHTML = '';

        if (files.length === 0) {
            const empty = document.createElement('div');
            empty.textContent = 'No files in this folder';
            empty.style.cssText = `
                text-align: center;
                padding: 40px;
                color: #999;
            `;
            fileList.appendChild(empty);
            return;
        }

        // Sort: folders first, then files
        const sorted = [...files].sort((a, b) => {
            if (a.is_folder && !b.is_folder) return -1;
            if (!a.is_folder && b.is_folder) return 1;
            return a.name.localeCompare(b.name);
        });

        sorted.forEach(file => {
            const item = document.createElement('div');
            item.style.cssText = `
                padding: 12px;
                border-bottom: 1px solid #f0f0f0;
                cursor: pointer;
                display: flex;
                align-items: center;
                gap: 12px;
                transition: background 0.2s;
            `;

            if (file.is_folder) {
                item.style.cursor = 'pointer';
                let existingPath = file.path;
                if (!existingPath) {
                    existingPath = file.name;
                } else {
                    existingPath = existingPath + '/' + file.name;
                }
                item.onclick = () => navigateToPath(existingPath);

            } else {
                item.onclick = () => selectFile(file);
                item.onmouseover = () => item.style.background = '#f8f9fa';
                item.onmouseout = () => {
                    if (selectedFile?.id !== file.id) {
                        item.style.background = 'transparent';
                    }
                };
            }

            // Icon
            const icon = document.createElement('span');
            icon.textContent = file.is_folder ? '📁' : '📄';
            icon.style.cssText = 'font-size: 24px;';

            // File info
            const info = document.createElement('div');
            info.style.cssText = 'flex: 1;';
            const name = document.createElement('div');
            name.textContent = file.name;
            name.style.cssText = `
                font-weight: ${file.is_folder ? '600' : '400'};
                color: ${file.is_folder ? '#333' : '#666'};
                margin-bottom: 4px;
            `;
            if (!file.is_folder) {
                const meta = document.createElement('div');
                meta.textContent = `${formatFileSize(file.size)} • ${file.mime || 'unknown'}`;
                meta.style.cssText = 'font-size: 12px; color: #999;';
                info.appendChild(meta);
            }
            info.insertBefore(name, info.firstChild);

            // Selection indicator
            if (selectedFile?.id === file.id) {
                item.style.background = '#e7f3ff';
                const check = document.createElement('span');
                check.textContent = '✓';
                check.style.cssText = 'color: #007bff; font-size: 20px; font-weight: bold;';
                item.appendChild(check);
            }

            item.appendChild(icon);
            item.appendChild(info);
            fileList.appendChild(item);
        });
    };

    const selectFile = (file) => {
        selectedFile = file;
        const selectBtn = document.getElementById('space-picker-select-btn');
        if (selectBtn) {
            selectBtn.disabled = false;
        }
        renderFileList(currentFiles || []);
    };

    let currentFiles = [];

    const navigateToPath = async (path) => {
        currentPath = path;
        updateBreadcrumb();

        const fileList = document.getElementById('space-picker-list');
        if (fileList) {
            fileList.innerHTML = '<div style="text-align: center; padding: 40px; color: #999;">Loading...</div>';
        }

        try {
            const url = new URL('/zz/api/core/space_file/list', window.location.origin);
            if (path) {
                url.searchParams.set('path', path);
            }

            const response = await fetch(url.toString(), {
                headers: {
                    'Autorization': spaceToken
                }
            });

            if (!response.ok) {
                if (response.status === 401) {
                    throw new Error('Unauthorized. Please authenticate.');
                }
                throw new Error(`Failed to load files: ${response.statusText}`);
            }

            const files = await response.json();
            currentFiles = files;
            selectedFile = null;
            const selectBtn = document.getElementById('space-picker-select-btn');
            if (selectBtn) {
                selectBtn.disabled = true;
            }
            renderFileList(files);
        } catch (error) {
            const fileList = document.getElementById('space-picker-list');
            if (fileList) {
                fileList.innerHTML = `
                    <div style="text-align: center; padding: 40px; color: #d32f2f;">
                        <div style="margin-bottom: 8px;">⚠️ Error loading files</div>
                        <div style="font-size: 14px;">${error.message}</div>
                    </div>
                `;
            }
            console.error('Error loading files:', error);
        }
    };

    const switchMode = (mode) => {
        currentMode = mode;
        const browseTab = document.getElementById('space-picker-tab-browse');
        const uploadTab = document.getElementById('space-picker-tab-upload');
        const breadcrumb = document.getElementById('space-picker-breadcrumb');
        const fileList = document.getElementById('space-picker-list');
        const uploadContainer = document.getElementById('space-picker-upload');
        const selectBtn = document.getElementById('space-picker-select-btn');
        const uploadBtn = document.getElementById('space-picker-upload-btn');

        if (mode === 'browse') {
            browseTab.style.color = '#007bff';
            browseTab.style.borderBottomColor = '#007bff';
            uploadTab.style.color = '#666';
            uploadTab.style.borderBottomColor = 'transparent';
            breadcrumb.style.display = 'flex';
            fileList.style.display = 'block';
            uploadContainer.style.display = 'none';
            selectBtn.style.display = 'block';
            uploadBtn.style.display = 'none';
        } else {
            browseTab.style.color = '#666';
            browseTab.style.borderBottomColor = 'transparent';
            uploadTab.style.color = '#007bff';
            uploadTab.style.borderBottomColor = '#007bff';
            breadcrumb.style.display = 'none';
            fileList.style.display = 'none';
            uploadContainer.style.display = 'block';
            selectBtn.style.display = 'none';
            uploadBtn.style.display = 'block';
        }
    };

    const addFilesToUpload = (newFiles) => {
        filesToUpload = [...filesToUpload, ...newFiles];
        renderUploadFilesList();
        const uploadBtn = document.getElementById('space-picker-upload-btn');
        if (uploadBtn) {
            uploadBtn.disabled = filesToUpload.length === 0;
        }
    };

    const removeFileFromUpload = (index) => {
        filesToUpload.splice(index, 1);
        renderUploadFilesList();
        const uploadBtn = document.getElementById('space-picker-upload-btn');
        if (uploadBtn) {
            uploadBtn.disabled = filesToUpload.length === 0;
        }
    };

    const renderUploadFilesList = () => {
        const filesList = document.getElementById('space-picker-upload-files');
        if (!filesList) return;

        filesList.innerHTML = '';

        if (filesToUpload.length === 0) {
            return;
        }

        filesToUpload.forEach((file, index) => {
            const item = document.createElement('div');
            item.style.cssText = `
                padding: 12px;
                border: 1px solid #e0e0e0;
                border-radius: 4px;
                margin-bottom: 8px;
                display: flex;
                align-items: center;
                gap: 12px;
                background: white;
            `;

            const icon = document.createElement('span');
            icon.textContent = '📄';
            icon.style.cssText = 'font-size: 24px;';

            const info = document.createElement('div');
            info.style.cssText = 'flex: 1;';
            const name = document.createElement('div');
            name.textContent = file.name;
            name.style.cssText = 'font-weight: 500; color: #333; margin-bottom: 4px;';
            const size = document.createElement('div');
            size.textContent = formatFileSize(file.size);
            size.style.cssText = 'font-size: 12px; color: #999;';
            info.appendChild(name);
            info.appendChild(size);

            const removeBtn = document.createElement('button');
            removeBtn.textContent = '×';
            removeBtn.style.cssText = `
                background: none;
                border: none;
                font-size: 24px;
                cursor: pointer;
                color: #999;
                padding: 0;
                width: 32px;
                height: 32px;
                line-height: 1;
            `;
            removeBtn.onclick = () => removeFileFromUpload(index);
            removeBtn.onmouseover = () => removeBtn.style.color = '#d32f2f';
            removeBtn.onmouseout = () => removeBtn.style.color = '#999';

            item.appendChild(icon);
            item.appendChild(info);
            item.appendChild(removeBtn);
            filesList.appendChild(item);
        });
    };

    const uploadFiles = async () => {
        if (filesToUpload.length === 0) return;

        const uploadBtn = document.getElementById('space-picker-upload-btn');
        if (uploadBtn) {
            uploadBtn.disabled = true;
            uploadBtn.textContent = 'Uploading...';
        }

        const filesList = document.getElementById('space-picker-upload-files');
        let successCount = 0;
        let errorCount = 0;
        const errors = [];

        // Upload files one at a time (backend only handles one file per request)
        for (let i = 0; i < filesToUpload.length; i++) {
            const file = filesToUpload[i];
            
            // Show progress
            if (filesList) {
                filesList.innerHTML = `
                    <div style="text-align: center; padding: 20px; color: #666;">
                        <div style="margin-bottom: 8px;">Uploading files...</div>
                        <div style="font-size: 12px; color: #999;">${i + 1} of ${filesToUpload.length}: ${file.name}</div>
                    </div>
                `;
            }

            try {
                const formData = new FormData();
                formData.append('files', file);
                formData.append('filename', file.name);

                const url = new URL('/zz/api/core/space_file/upload', window.location.origin);
                if (currentPath) {
                    url.searchParams.set('path', currentPath);
                }

                const response = await fetch(url.toString(), {
                    method: 'POST',
                    headers: {
                        'Autorization': spaceToken
                    },
                    body: formData
                });

                if (!response.ok) {
                    if (response.status === 401) {
                        throw new Error('Unauthorized. Please authenticate.');
                    }
                    const errorText = await response.text();
                    throw new Error(`${file.name}: ${errorText || response.statusText}`);
                }

                successCount++;
            } catch (error) {
                errorCount++;
                errors.push(error.message);
                console.error(`Error uploading ${file.name}:`, error);
            }
        }

        // Show final result
        if (filesList) {
            if (errorCount === 0) {
                filesList.innerHTML = `
                    <div style="text-align: center; padding: 20px; color: #28a745;">
                        <div style="margin-bottom: 8px;">✓ All files uploaded successfully!</div>
                        <div style="font-size: 12px; color: #666;">${successCount} file(s) uploaded</div>
                    </div>
                `;
            } else if (successCount === 0) {
                filesList.innerHTML = `
                    <div style="text-align: center; padding: 20px; color: #d32f2f;">
                        <div style="margin-bottom: 8px;">⚠️ Upload failed</div>
                        <div style="font-size: 14px; text-align: left; margin-top: 12px;">
                            ${errors.map(e => `<div style="margin-bottom: 4px;">• ${e}</div>`).join('')}
                        </div>
                    </div>
                `;
            } else {
                filesList.innerHTML = `
                    <div style="text-align: center; padding: 20px;">
                        <div style="margin-bottom: 8px; color: #ff9800;">⚠️ Partial success</div>
                        <div style="font-size: 12px; color: #666; margin-bottom: 8px;">
                            ${successCount} uploaded, ${errorCount} failed
                        </div>
                        <div style="font-size: 14px; text-align: left; margin-top: 12px; color: #d32f2f;">
                            ${errors.map(e => `<div style="margin-bottom: 4px;">• ${e}</div>`).join('')}
                        </div>
                    </div>
                `;
            }
        }

        // Clear files and reset after 2 seconds if all successful
        if (errorCount === 0) {
            setTimeout(() => {
                filesToUpload = [];
                renderUploadFilesList();
                if (uploadBtn) {
                    uploadBtn.disabled = true;
                    uploadBtn.textContent = 'Upload';
                }
                // Optionally switch to browse mode and refresh
                if (currentMode === 'upload') {
                    switchMode('browse');
                    navigateToPath(currentPath);
                }
            }, 2000);
        } else {
            // Keep files if there were errors, allow retry
            if (uploadBtn) {
                uploadBtn.disabled = false;
                uploadBtn.textContent = 'Upload';
            }
        }
    };

    const closeModal = () => {
        if (modal && modal.parentNode) {
            modal.parentNode.removeChild(modal);
        }
        modal = null;
        selectedFile = null;
        currentPath = '';
        currentFiles = [];
        filesToUpload = [];
        currentMode = 'browse';
    };

    return {
        showModal: (onSelect) => {
            if (modal) {
                closeModal();
            }
            onSelectCallback = onSelect;
            modal = createModal();
            document.body.appendChild(modal);
            switchMode('browse');
            navigateToPath('');
        },
        close: closeModal
    };
}




window.spaceRedirrectToAuth = spaceRedirrectToAuth;
window.spaceGetToken = spaceGetToken;
window.spaceFilePicker = spaceFilePicker;

console.log("libspace.js/end");
