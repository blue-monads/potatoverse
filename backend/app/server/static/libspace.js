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
        title.textContent = 'Select File';
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

        // Breadcrumb
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

        // File list container
        const fileList = document.createElement('div');
        fileList.id = 'space-picker-list';
        fileList.style.cssText = `
            flex: 1;
            overflow-y: auto;
            padding: 8px;
            min-height: 300px;
        `;

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
            if (selectedFile && onSelectCallback) {
                onSelectCallback(selectedFile);
                closeModal();
            }
        };
        selectBtn.disabled = true;
        footer.appendChild(cancelBtn);
        footer.appendChild(selectBtn);

        content.appendChild(header);
        content.appendChild(breadcrumb);
        content.appendChild(fileList);
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
            const url = new URL('/zz/api/core/signed_file/list', window.location.origin);
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

    const closeModal = () => {
        if (modal && modal.parentNode) {
            modal.parentNode.removeChild(modal);
        }
        modal = null;
        selectedFile = null;
        currentPath = '';
        currentFiles = [];
    };

    return {
        showModal: (onSelect) => {
            if (modal) {
                closeModal();
            }
            onSelectCallback = onSelect;
            modal = createModal();
            document.body.appendChild(modal);
            navigateToPath('');
        },
        close: closeModal
    };
}




window.spaceRedirrectToAuth = spaceRedirrectToAuth;
window.spaceGetToken = spaceGetToken;
window.spaceFilePicker = spaceFilePicker;

console.log("libspace.js/end");
