"use client";
import React, { useEffect, useRef, useState } from 'react';
import {
    Folder,
    File,
    Download,
    Trash2,
    Upload,
    ArrowLeft,
    Search,
    Grid3X3,
    List,
    Plus
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { listSpaceFiles, SpaceFile, deleteSpaceFile, downloadSpaceFile, uploadSpaceFile, createSpaceFolder, createPresignedUploadURL, PresignedUploadResponse } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { useGApp } from '@/hooks';
import BigSearchBar from '@/contain/compo/BigSearchBar';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const gapp = useGApp();
    const { modal } = gapp;

    if (!installId) {
        return (
            <WithAdminBodyLayout Icon={Folder} name="Space Files" description="Select a space to view files">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <Folder className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                        <p className="text-gray-500">No space selected</p>
                        <button
                            onClick={() => router.back()}
                            className="mt-4 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600"
                        >
                            Go Back
                        </button>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    return <FileManager installId={parseInt(installId)} />;
}

interface FileManagerProps {
    installId: number;
}

const FileManager = ({ installId }: FileManagerProps) => {
    const [currentPath, setCurrentPath] = useState('');
    const [searchTerm, setSearchTerm] = useState('');
    const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
    const [uploading, setUploading] = useState(false);
    const [newFolderName, setNewFolderName] = useState('');
    const fileInputRef = useRef<HTMLInputElement>(null);
    const gapp = useGApp();
    const { modal } = gapp;

    const loader = useSimpleDataLoader<SpaceFile[]>({
        loader: () => listSpaceFiles(installId, currentPath),
        ready: gapp.isInitialized,
        dependencies: [currentPath, searchTerm],
    });

    const files = loader.data || [];
    const filteredFiles = files.filter(file => {
        const matchesPath = file.path === currentPath;
        const matchesSearch = file.name.toLowerCase().includes(searchTerm.toLowerCase());
        return matchesPath && matchesSearch;
    });

    const folders = filteredFiles.filter(file => file.is_folder);
    const fileItems = filteredFiles.filter(file => !file.is_folder);

    const breadcrumbs = currentPath.split('/').filter(Boolean);

    const handleFolderClick = (folder: SpaceFile) => {
        const newPath = currentPath ? `${currentPath}/${folder.name}` : folder.name;
        setCurrentPath(newPath);
    };

    const handleBackClick = () => {
        const pathParts = currentPath.split('/');
        pathParts.pop();
        setCurrentPath(pathParts.join('/'));
    };

    const handleFileDownload = async (file: SpaceFile) => {
        try {
            const response = await downloadSpaceFile(installId, file.id);
            const url = window.URL.createObjectURL(new Blob([response.data]));
            const link = document.createElement('a');
            link.href = url;
            link.setAttribute('download', file.name);
            document.body.appendChild(link);
            link.click();
            link.remove();
            window.URL.revokeObjectURL(url);
        } catch (error) {
            console.error('Download failed:', error);
        }
    };

    const handleFileDelete = async (file: SpaceFile) => {
        if (!confirm(`Are you sure you want to delete "${file.name}"?`)) return;

        try {
            await deleteSpaceFile(installId, file.id);
            loader.reload();
        } catch (error) {
            console.error('Delete failed:', error);
        }
    };

    const handleCreateFolder = async () => {
        if (!newFolderName.trim()) return;

        try {
            await createSpaceFolder(installId, newFolderName.trim(), currentPath);
            setNewFolderName('');
            modal.closeModal();
            loader.reload();
        } catch (error) {
            console.error('Create folder failed:', error);
        }
    };

    const UploadModalContent = () => {
        const [uploadMode, setUploadMode] = useState<'regular' | 'presigned'>('regular');
        const [presignedFileName, setPresignedFileName] = useState('');
        const [presignedData, setPresignedData] = useState<PresignedUploadResponse | null>(null);
        const [exampleTab, setExampleTab] = useState<'curl' | 'javascript' | 'python' | 'browser'>('curl');

        const handleGeneratePresignedURL = async () => {
            if (!presignedFileName.trim()) return;

            try {
                const response = await createPresignedUploadURL(installId, presignedFileName.trim(), currentPath, 3600);
                setPresignedData(response.data);
            } catch (error) {
                console.error('Generate presigned URL failed:', error);
            }
        };

        return (
            <div className="space-y-4">
                {/* Upload Mode Selection */}
                <div className="flex gap-2 p-1 bg-gray-100 rounded-lg">
                    <button
                        onClick={() => setUploadMode('regular')}
                        className={`flex-1 px-4 py-2 rounded-lg font-medium transition-all ${
                            uploadMode === 'regular'
                                ? 'bg-white text-blue-600 shadow-sm'
                                : 'text-gray-600 hover:text-gray-900'
                        }`}
                    >
                        Regular Upload
                    </button>
                    <button
                        onClick={() => setUploadMode('presigned')}
                        className={`flex-1 px-4 py-2 rounded-lg font-medium transition-all ${
                            uploadMode === 'presigned'
                                ? 'bg-white text-purple-600 shadow-sm'
                                : 'text-gray-600 hover:text-gray-900'
                        }`}
                    >
                        Presigned Upload
                    </button>
                </div>

                    {/* Regular Upload Mode */}
                    {uploadMode === 'regular' && (
                        <div className="space-y-4">
                            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                                <h3 className="font-semibold text-blue-900 mb-2">Regular Upload</h3>
                                <p className="text-sm text-blue-800">
                                    Upload files directly from your device. Files will be uploaded immediately to the current directory.
                                </p>
                            </div>

                            <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center hover:border-blue-400 transition-colors">
                                <Upload className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                                <p className="text-gray-600 mb-2">Click to select files or drag and drop</p>
                                <p className="text-xs text-gray-500">Multiple files supported</p>
                                <input
                                    ref={fileInputRef}
                                    type="file"
                                    multiple
                                    onChange={async (e) => {
                                        const files = e.target.files;
                                        if (!files || files.length === 0) return;

                                        setUploading(true);
                                        try {
                                            for (const file of Array.from(files)) {
                                                await uploadSpaceFile(installId, file, currentPath);
                                            }
                                            loader.reload();
                                            modal.closeModal();
                                        } catch (error) {
                                            console.error('Upload failed:', error);
                                            alert('Upload failed: ' + error);
                                        } finally {
                                            setUploading(false);
                                            if (fileInputRef.current) {
                                                fileInputRef.current.value = '';
                                            }
                                        }
                                    }}
                                    className="hidden"
                                />
                                <button
                                    onClick={() => fileInputRef.current?.click()}
                                    disabled={uploading}
                                    className="mt-4 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
                                >
                                    {uploading ? 'Uploading...' : 'Select Files'}
                                </button>
                            </div>
                        </div>
                    )}

                    {/* Presigned Upload Mode */}
                    {uploadMode === 'presigned' && (
                        <div className="space-y-4">
                            <div className="bg-purple-50 border border-purple-200 rounded-lg p-4">
                                <h3 className="font-semibold text-purple-900 mb-2">Presigned Upload</h3>
                                <p className="text-sm text-purple-800">
                                    Generate a temporary token that can be used to upload files without authentication. 
                                    Perfect for third-party integrations, CLI tools, or sharing upload capabilities securely.
                                </p>
                            </div>

                            <div>
                                <label className="block text-sm font-medium text-gray-700 mb-2">
                                    File Name
                                </label>
                                <input
                                    type="text"
                                    value={presignedFileName}
                                    onChange={(e) => setPresignedFileName(e.target.value)}
                                    placeholder="e.g., document.pdf"
                                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500"
                                />
                                <p className="text-xs text-gray-500 mt-1">
                                    The uploaded file must match this exact filename
                                </p>
                            </div>

                            <button
                                onClick={handleGeneratePresignedURL}
                                disabled={!presignedFileName.trim()}
                                className="w-full px-4 py-2 bg-purple-500 text-white rounded-lg hover:bg-purple-600 disabled:opacity-50"
                            >
                                Generate Presigned Token
                            </button>

                            {presignedData && (
                                <div className="space-y-4 pt-4 border-t border-gray-200">
                                    <div className="bg-green-50 border border-green-200 rounded-lg p-3">
                                        <p className="text-sm text-green-800 font-medium">
                                            ✓ Token generated successfully! Expires in {presignedData.expiry} seconds
                                        </p>
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-2">
                                            Presigned Token
                                        </label>
                                        <div className="relative">
                                            <input
                                                type="text"
                                                value={presignedData.presigned_token}
                                                readOnly
                                                className="w-full px-3 py-2 bg-gray-50 border border-gray-300 rounded-lg font-mono text-xs"
                                            />
                                            <button
                                                onClick={() => {
                                                    navigator.clipboard.writeText(presignedData.presigned_token);
                                                }}
                                                className="absolute right-2 top-2 px-2 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded"
                                            >
                                                Copy
                                            </button>
                                        </div>
                                    </div>

                                    <div>
                                        <label className="block text-sm font-medium text-gray-700 mb-2">
                                            Upload URL
                                        </label>
                                        <div className="relative">
                                            <input
                                                type="text"
                                                value={`${window.location.origin}${presignedData.upload_url}`}
                                                readOnly
                                                className="w-full px-3 py-2 bg-gray-50 border border-gray-300 rounded-lg font-mono text-xs"
                                            />
                                            <button
                                                onClick={() => {
                                                    navigator.clipboard.writeText(`${window.location.origin}${presignedData.upload_url}`);
                                                }}
                                                className="absolute right-2 top-2 px-2 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded"
                                            >
                                                Copy
                                            </button>
                                        </div>
                                    </div>

                                    <div className="bg-gray-50 rounded-lg p-4">
                                        <h4 className="font-semibold text-gray-900 mb-3">Upload Examples</h4>
                                        
                                        {/* Example Tabs */}
                                        <div className="flex gap-1 mb-3 overflow-x-auto">
                                            <button
                                                onClick={() => setExampleTab('curl')}
                                                className={`px-3 py-1.5 text-xs font-medium rounded transition-colors whitespace-nowrap ${
                                                    exampleTab === 'curl'
                                                        ? 'bg-gray-900 text-white'
                                                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                                                }`}
                                            >
                                                cURL
                                            </button>
                                            <button
                                                onClick={() => setExampleTab('javascript')}
                                                className={`px-3 py-1.5 text-xs font-medium rounded transition-colors whitespace-nowrap ${
                                                    exampleTab === 'javascript'
                                                        ? 'bg-gray-900 text-white'
                                                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                                                }`}
                                            >
                                                JavaScript
                                            </button>
                                            <button
                                                onClick={() => setExampleTab('python')}
                                                className={`px-3 py-1.5 text-xs font-medium rounded transition-colors whitespace-nowrap ${
                                                    exampleTab === 'python'
                                                        ? 'bg-gray-900 text-white'
                                                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                                                }`}
                                            >
                                                Python
                                            </button>
                                            <button
                                                onClick={() => setExampleTab('browser')}
                                                className={`px-3 py-1.5 text-xs font-medium rounded transition-colors whitespace-nowrap ${
                                                    exampleTab === 'browser'
                                                        ? 'bg-gray-900 text-white'
                                                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                                                }`}
                                            >
                                                Browser
                                            </button>
                                        </div>

                                        {/* Example Content */}
                                        <div>
                                            {exampleTab === 'curl' && (
                                                <div>
                                                    <pre className="bg-gray-900 text-green-400 p-3 rounded-lg text-xs overflow-x-auto">
{`curl -X POST \\
  "${window.location.origin}${presignedData.upload_url}" \\
  -F "file=@${presignedFileName}"`}
                                                    </pre>
                                                    <button
                                                        onClick={() => {
                                                            navigator.clipboard.writeText(`curl -X POST "${window.location.origin}${presignedData.upload_url}" -F "file=@${presignedFileName}"`);
                                                        }}
                                                        className="mt-2 text-xs text-purple-600 hover:text-purple-800"
                                                    >
                                                        Copy Command
                                                    </button>
                                                </div>
                                            )}

                                            {exampleTab === 'javascript' && (
                                                <div>
                                                    <pre className="bg-gray-900 text-green-400 p-3 rounded-lg text-xs overflow-x-auto">
{`const formData = new FormData();
formData.append('file', fileInput.files[0]);

fetch('${window.location.origin}${presignedData.upload_url}', {
  method: 'POST',
  body: formData
}).then(res => res.json())
  .then(data => console.log(data));`}
                                                    </pre>
                                                    <button
                                                        onClick={() => {
                                                            navigator.clipboard.writeText(`const formData = new FormData();\nformData.append('file', fileInput.files[0]);\n\nfetch('${window.location.origin}${presignedData.upload_url}', {\n  method: 'POST',\n  body: formData\n}).then(res => res.json())\n  .then(data => console.log(data));`);
                                                        }}
                                                        className="mt-2 text-xs text-purple-600 hover:text-purple-800"
                                                    >
                                                        Copy Code
                                                    </button>
                                                </div>
                                            )}

                                            {exampleTab === 'python' && (
                                                <div>
                                                    <pre className="bg-gray-900 text-green-400 p-3 rounded-lg text-xs overflow-x-auto">
{`import requests

with open('${presignedFileName}', 'rb') as f:
    files = {'file': f}
    response = requests.post(
        '${window.location.origin}${presignedData.upload_url}',
        files=files
    )
print(response.json())`}
                                                    </pre>
                                                    <button
                                                        onClick={() => {
                                                            navigator.clipboard.writeText(`import requests\n\nwith open('${presignedFileName}', 'rb') as f:\n    files = {'file': f}\n    response = requests.post(\n        '${window.location.origin}${presignedData.upload_url}',\n        files=files\n    )\nprint(response.json())`);
                                                        }}
                                                        className="mt-2 text-xs text-purple-600 hover:text-purple-800"
                                                    >
                                                        Copy Code
                                                    </button>
                                                </div>
                                            )}

                                            {exampleTab === 'browser' && (
                                                <div className="space-y-3">
                                                    <p className="text-sm text-gray-600">
                                                        Share this link with users to upload via web browser:
                                                    </p>
                                                    <div className="relative">
                                                        <input
                                                            type="text"
                                                            value={`${window.location.origin}/zz/pages/presigned/file?presigned-key=${presignedData.presigned_token}`}
                                                            readOnly
                                                            className="w-full px-3 py-2 bg-gray-100 border border-gray-300 rounded-lg font-mono text-xs"
                                                        />
                                                        <button
                                                            onClick={() => {
                                                                navigator.clipboard.writeText(`${window.location.origin}/zz/pages/presigned/file?presigned-key=${presignedData.presigned_token}`);
                                                            }}
                                                            className="absolute right-2 top-2 px-2 py-1 text-xs bg-gray-200 hover:bg-gray-300 rounded"
                                                        >
                                                            Copy
                                                        </button>
                                                    </div>
                                                    <a
                                                        href={`/zz/pages/presigned/file?presigned-key=${presignedData.presigned_token}`}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        className="inline-block text-xs text-purple-600 hover:text-purple-800"
                                                    >
                                                        Open Upload Page →
                                                    </a>
                                                </div>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            )}
                        </div>
                    )}

                <div className="flex justify-end gap-2 pt-4 border-t border-gray-200">
                    <button
                        onClick={() => {
                            modal.closeModal();
                            setPresignedFileName('');
                            setPresignedData(null);
                        }}
                        className="px-4 py-2 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200"
                    >
                        Close
                    </button>
                </div>
            </div>
        );
    };

    const showUploadModal = () => {
        modal.openModal({
            title: "Upload Files",
            content: <UploadModalContent />,
            size: "lg",
        });
    };

    return (
        <WithAdminBodyLayout
            Icon={Folder}
            name="Space Files"
            description={`Managing files for install ID ${installId}`}
            rightContent={
                <div className="flex items-center gap-2">
                    <button
                        onClick={() => {
                            setNewFolderName('');
                            modal.openModal({
                                title: "Create New Folder",
                                content: (
                                    <div className="space-y-4">
                                        <div>
                                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                                Folder Name
                                            </label>
                                            <input
                                                type="text"
                                                value={newFolderName}
                                                onChange={(e) => setNewFolderName(e.target.value)}
                                                placeholder="Enter folder name"
                                                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                                                autoFocus
                                                onKeyDown={(e) => {
                                                    if (e.key === 'Enter') {
                                                        handleCreateFolder();
                                                    } else if (e.key === 'Escape') {
                                                        modal.closeModal();
                                                        setNewFolderName('');
                                                    }
                                                }}
                                            />
                                        </div>
                                        <div className="flex justify-end gap-2">
                                            <button
                                                onClick={() => {
                                                    modal.closeModal();
                                                    setNewFolderName('');
                                                }}
                                                className="px-4 py-2 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200"
                                            >
                                                Cancel
                                            </button>
                                            <button
                                                onClick={handleCreateFolder}
                                                disabled={!newFolderName.trim()}
                                                className="px-4 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 disabled:opacity-50"
                                            >
                                                Create Folder
                                            </button>
                                        </div>
                                    </div>
                                ),
                                size: "md",
                                onClose: () => setNewFolderName('')
                            });
                        }}
                        className="flex items-center gap-2 px-3 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600"
                    >
                        <Plus className="w-4 h-4" />
                        New Folder
                    </button>
                    <button
                        onClick={showUploadModal}
                        disabled={uploading}
                        className="flex items-center gap-2 px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
                    >
                        <Upload className="w-4 h-4" />
                        {uploading ? 'Uploading...' : 'Upload Files'}
                    </button>
                </div>
            }
        >

            <div className="card m-4 p-4 flex flex-col gap-4">

                {/* Search and View Controls */}
                <div className="flex items-center justify-between bg-white">

                    <div className='flex w-full'>
                        <BigSearchBar
                            searchText={searchTerm}
                            setSearchText={setSearchTerm}
                            placeholder="Search files..."
                            onSearchButtonClick={() => loader.reload()}
                            className="w-full"
                        />

                    </div>

                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setViewMode('list')}
                            className={`p-2 rounded ${viewMode === 'list' ? 'bg-blue-100 text-blue-600' : 'text-gray-400'}`}
                        >
                            <List className="w-4 h-4" />
                        </button>
                        <button
                            onClick={() => setViewMode('grid')}
                            className={`p-2 rounded ${viewMode === 'grid' ? 'bg-blue-100 text-blue-600' : 'text-gray-400'}`}
                        >
                            <Grid3X3 className="w-4 h-4" />
                        </button>
                    </div>
                </div>

                {/* Breadcrumbs */}
                <nav className="flex items-center space-x-2 text-sm rounded-lg bg-white p-4 border border-gray-200">
                    <button
                        onClick={() => setCurrentPath('')}
                        className="text-blue-500 hover:text-blue-700"
                    >
                        Root
                    </button>
                    {breadcrumbs.map((crumb, index) => (
                        <React.Fragment key={index}>
                            <span className="text-gray-400">/</span>
                            <button
                                onClick={() => {
                                    const path = breadcrumbs.slice(0, index + 1).join('/');
                                    setCurrentPath(path);
                                }}
                                className="text-blue-500 hover:text-blue-700"
                            >
                                {crumb}
                            </button>
                        </React.Fragment>
                    ))}
                </nav>

                {/* Back Button */}
                {currentPath && (
                    <div className="mb-4">
                        <button
                            onClick={handleBackClick}
                            className="flex items-center gap-2 text-blue-500 hover:text-blue-700"
                        >
                            <ArrowLeft className="w-4 h-4" />
                            Back
                        </button>
                    </div>
                )}

                {/* File List */}
                {loader.loading ? (
                    <div className="flex items-center justify-center h-64">
                        <div className="text-center">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                            <p className="text-gray-500">Loading files...</p>
                        </div>
                    </div>
                ) : filteredFiles.length === 0 ? (
                    <div className="flex items-center justify-center h-64">
                        <div className="text-center">
                            <Folder className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                            <p className="text-gray-500">No files found</p>
                        </div>
                    </div>
                ) : (
                    <div className={viewMode === 'grid' ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4' : 'space-y-2'}>
                        {/* Folders */}
                        {folders.map((folder) => (
                            <FileItem
                                key={folder.id}
                                file={folder}
                                viewMode={viewMode}
                                onDoubleClick={() => handleFolderClick(folder)}
                                onDownload={() => { }}
                                onDelete={() => handleFileDelete(folder)}
                            />
                        ))}

                        {/* Files */}
                        {fileItems.map((file) => (
                            <FileItem
                                key={file.id}
                                file={file}
                                viewMode={viewMode}
                                onDoubleClick={() => handleFileDownload(file)}
                                onDownload={() => handleFileDownload(file)}
                                onDelete={() => handleFileDelete(file)}
                            />
                        ))}
                    </div>
                )}

            </div>

        </WithAdminBodyLayout>
    );
};

interface FileItemProps {
    file: SpaceFile;
    viewMode: 'list' | 'grid';
    onDoubleClick: () => void;
    onDownload: () => void;
    onDelete: () => void;
}

const FileItem = ({ file, viewMode, onDoubleClick, onDownload, onDelete }: FileItemProps) => {
    const [showActions, setShowActions] = useState(false);

    if (viewMode === 'grid') {
        return (
            <div
                className="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer group relative"
                onDoubleClick={onDoubleClick}
                onMouseEnter={() => setShowActions(true)}
                onMouseLeave={() => setShowActions(false)}
            >
                <div className="flex flex-col items-center text-center">
                    <div className="w-12 h-12 mb-2 flex items-center justify-center">
                        {file.is_folder ? (
                            <Folder className="w-8 h-8 text-blue-500" />
                        ) : (
                            <File className="w-8 h-8 text-gray-500" />
                        )}
                    </div>
                    <h3 className="text-sm font-medium text-gray-900 truncate w-full">{file.name}</h3>
                    <p className="text-xs text-gray-500">
                        {file.is_folder ? 'Folder' : formatFileSize(file.size)}
                    </p>
                </div>

                {showActions && !file.is_folder && (
                    <div className="absolute top-2 right-2 flex gap-1">
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onDownload();
                            }}
                            className="p-1 bg-white rounded shadow-sm hover:bg-gray-100"
                        >
                            <Download className="w-3 h-3" />
                        </button>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onDelete();
                            }}
                            className="p-1 bg-white rounded shadow-sm hover:bg-gray-100 text-red-500"
                        >
                            <Trash2 className="w-3 h-3" />
                        </button>
                    </div>
                )}
            </div>
        );
    }

    return (
        <div
            className="flex items-center p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer group"
            onDoubleClick={onDoubleClick}
            onMouseEnter={() => setShowActions(true)}
            onMouseLeave={() => setShowActions(false)}
        >
            <div className="flex-shrink-0 mr-3">
                {file.is_folder ? (
                    <Folder className="w-5 h-5 text-blue-500" />
                ) : (
                    <File className="w-5 h-5 text-gray-500" />
                )}
            </div>

            <div className="flex-1 min-w-0">
                <h3 className="text-sm font-medium text-gray-900 truncate">{file.name}</h3>
                <p className="text-xs text-gray-500">
                    {file.is_folder ? 'Folder' : `${formatFileSize(file.size)} • ${formatDate(file.created_at)}`}
                </p>
            </div>

            {showActions && !file.is_folder && (
                <div className="flex items-center gap-1">
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            onDownload();
                        }}
                        className="p-1 hover:bg-gray-200 rounded"
                    >
                        <Download className="w-4 h-4" />
                    </button>
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            onDelete();
                        }}
                        className="p-1 hover:bg-gray-200 rounded text-red-500"
                    >
                        <Trash2 className="w-4 h-4" />
                    </button>
                </div>
            )}
        </div>
    );
};

const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};

const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
};
