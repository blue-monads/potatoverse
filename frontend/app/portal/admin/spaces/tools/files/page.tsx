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
    Plus,
    Edit,
    ChevronRight,
    ChevronLeft,
    Key,
    Check
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { listSpaceFiles, SpaceFile, deleteSpaceFile, downloadSpaceFile, uploadSpaceFile, createSpaceFolder, createPresignedUploadURL, PresignedUploadResponse, updateSpaceFileContent } from '@/lib';
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

    const handleFileEdit = async (file: SpaceFile) => {
        try {
            const response = await downloadSpaceFile(installId, file.id);
            if (!(response.data instanceof Blob)) {
                throw new Error('Expected blob response but got something else');
            }
            let content = await response.data.text();

            modal.openModal({
                title: `Edit ${file.name}`,
                content: (
                    <FileEditor
                        installId={installId}
                        file={file}
                        initialContent={content}
                        currentPath={currentPath}
                        onSave={() => {
                            loader.reload();
                            modal.closeModal();
                        }}
                        onCancel={() => modal.closeModal()}
                    />
                ),
                size: 'xl'
            });
        } catch (error) {
            console.error('Failed to load file for editing:', error);
            alert('Failed to load file for editing. The file might be too large or not a text file.');
        }
    };

    const handleCreateFolder = async (folderName: string) => {
        if (!folderName.trim()) return;
        try {
            await createSpaceFolder(installId, folderName.trim(), currentPath);
            modal.closeModal();
            loader.reload();
        } catch (error) {
            console.error('Create folder failed:', error);
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Folder}
            name="Files"
            description={`Managing source files for spaces`}
            variant="none"
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
                placeholder="Search files..."
                onSearchButtonClick={() => loader.reload()}
                rightContent={
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => {
                                modal.openModal({
                                    title: "Create New Folder",
                                    content: <CreateFolderModalContent handleCreateFolder={handleCreateFolder} modal={modal} />,
                                    size: "md",
                                });
                            }}
                            className="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 shadow-sm font-semibold transition-all"
                        >
                            <Plus className="w-4 h-4 text-green-500" />
                            New Folder
                        </button>
                        <button
                            onClick={() => {
                                modal.openModal({
                                    title: "Upload Files",
                                    content: <UploadModalContent installId={installId} currentPath={currentPath} loader={loader} modal={modal} fileInputRef={fileInputRef} setUploading={setUploading} uploading={uploading} />,
                                    size: "lg",
                                });
                            }}
                            disabled={uploading}
                            className="flex items-center gap-2 px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 shadow-sm font-semibold transition-all"
                        >
                            <Upload className="w-4 h-4" />
                            {uploading ? 'Uploading...' : 'Upload'}
                        </button>
                    </div>
                }
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full flex flex-col gap-6">
                {/* Reference UI Title Header */}
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2 text-sm text-gray-400">
                            <Folder className="w-4 h-4" />
                            <span>Files</span>
                        </div>
                    </div>
                    <div className="flex items-center gap-2">
                        <button
                            onClick={() => setViewMode('list')}
                            className={`p-2 rounded-lg border transition-all ${viewMode === 'list' ? 'bg-blue-50 border-blue-200 text-blue-600' : 'bg-white border-gray-200 text-gray-400 hover:bg-gray-50'}`}
                        >
                            <List className="w-4 h-4" />
                        </button>
                        <button
                            onClick={() => setViewMode('grid')}
                            className={`p-2 rounded-lg border transition-all ${viewMode === 'grid' ? 'bg-blue-50 border-blue-200 text-blue-600' : 'bg-white border-gray-200 text-gray-400 hover:bg-gray-50'}`}
                        >
                            <Grid3X3 className="w-4 h-4" />
                        </button>
                    </div>
                </div>

                {/* Explorer Container */}
                <div className={`${viewMode === 'grid' ? '' : 'bg-white rounded-xl overflow-hidden shadow-sm border border-gray-200'}`}>
                    {viewMode === 'list' && (
                        <div className="bg-gray-50 px-6 py-4 flex items-center gap-3 border-b border-gray-200">
                            <div className="w-8 h-8 bg-blue-50 rounded flex items-center justify-center">
                                <Folder className="w-4 h-4 text-blue-500" />
                            </div>
                            <h2 className="text-gray-900 font-bold tracking-wide">
                                {currentPath ? `Root / ${currentPath.split('/').join(' / ')}` : 'Root'}
                            </h2>
                        </div>
                    )}
                    {viewMode === 'grid' ? (
                        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
                            {currentPath && (
                                <div
                                    onClick={handleBackClick}
                                    className="p-4 border border-gray-200 rounded-xl bg-white hover:bg-gray-50 cursor-pointer flex items-center gap-3 transition-all"
                                >
                                    <div className="w-10 h-10 bg-gray-50 rounded flex items-center justify-center">
                                        <ArrowLeft className="w-5 h-5 text-gray-400" />
                                    </div>
                                    <span className="font-semibold text-gray-500">Go Back</span>
                                </div>
                            )}
                            {[...folders, ...fileItems].map((file) => (
                                <div
                                    key={file.id}
                                    className="p-4 border border-gray-200 rounded-xl bg-white hover:border-blue-400 hover:shadow-md cursor-pointer group relative transition-all"
                                    onClick={() => file.is_folder ? handleFolderClick(file) : handleFileDownload(file)}
                                >
                                    <div className="flex flex-col items-center text-center">
                                        <div className="w-12 h-12 mb-3 flex items-center justify-center">
                                            {file.is_folder ? (
                                                <Folder className="w-10 h-10 text-blue-500" />
                                            ) : (
                                                <File className="w-10 h-10 text-gray-400" />
                                            )}
                                        </div>
                                        <h3 className="text-sm font-bold text-gray-900 truncate w-full">{file.name}</h3>
                                        <p className="text-[10px] uppercase tracking-wider font-bold text-gray-400 mt-1">
                                            {file.is_folder ? 'Folder' : formatFileSize(file.size)}
                                        </p>
                                    </div>

                                    {/* Grid Actions */}
                                    <div className="absolute top-2 right-2 hidden group-hover:flex gap-1.5">
                                        {!file.is_folder && (
                                            <button
                                                onClick={(e) => { e.stopPropagation(); handleFileEdit(file); }}
                                                className="p-1.5 bg-white border border-gray-200 rounded shadow-sm hover:bg-blue-50 text-gray-500 hover:text-blue-600 transition-all"
                                            >
                                                <Edit className="w-3 h-3" />
                                            </button>
                                        )}
                                        <button
                                            onClick={(e) => { e.stopPropagation(); handleFileDelete(file); }}
                                            className="p-1.5 bg-white border border-gray-200 rounded shadow-sm hover:bg-red-50 text-gray-400 hover:text-red-500 transition-all"
                                        >
                                            <Trash2 className="w-3 h-3" />
                                        </button>
                                    </div>
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div className="divide-y divide-gray-100">
                            {loader.loading ? (
                                <div className="p-20 text-center">
                                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                                    <p className="text-gray-500 text-sm">Loading source files...</p>
                                </div>
                            ) : (
                                <>
                                    {currentPath && (
                                        <div
                                            onClick={handleBackClick}
                                            className="px-6 py-4 hover:bg-gray-50 cursor-pointer group flex items-center transition-colors"
                                        >
                                            <div className="w-6 h-6 mr-4 flex items-center justify-center">
                                                <ChevronLeft className="w-4 h-4 text-gray-400 group-hover:text-blue-500" />
                                            </div>
                                            <span className="text-gray-500 group-hover:text-blue-600 text-sm font-medium">.. (Parent directory)</span>
                                        </div>
                                    )}
                                    {[...folders, ...fileItems].map(file => (
                                        <div
                                            key={file.id}
                                            className="px-6 py-4 hover:bg-gray-50 cursor-pointer group flex items-center transition-colors"
                                            onClick={() => file.is_folder ? handleFolderClick(file) : handleFileDownload(file)}
                                        >
                                            <div className="w-6 h-6 mr-4 flex items-center justify-center">
                                                {file.is_folder ? (
                                                    <Folder className="w-4 h-4 text-blue-400 group-hover:text-blue-500" />
                                                ) : (
                                                    <File className="w-4 h-4 text-gray-400 group-hover:text-gray-600" />
                                                )}
                                            </div>
                                            <span className={`flex-1 text-sm font-medium ${file.is_folder ? 'text-gray-700 group-hover:text-blue-600' : 'text-gray-600 group-hover:text-gray-900'}`}>
                                                {file.name}
                                            </span>
                                            <div className="flex items-center gap-6">
                                                {/* List Actions */}
                                                <div className="hidden group-hover:flex items-center gap-2 mr-2">
                                                    {!file.is_folder && (
                                                        <button
                                                            onClick={(e) => { e.stopPropagation(); handleFileEdit(file); }}
                                                            className="p-1.5 bg-white border border-gray-200 rounded hover:bg-blue-50 hover:border-blue-200 text-gray-500 hover:text-blue-600 transition-all"
                                                            title="Edit file"
                                                        >
                                                            <Edit className="w-3.5 h-3.5" />
                                                        </button>
                                                    )}
                                                    <button
                                                        onClick={(e) => { e.stopPropagation(); handleFileDownload(file); }}
                                                        className="p-1.5 bg-white border border-gray-200 rounded hover:bg-blue-50 hover:border-blue-200 text-gray-500 hover:text-blue-600 transition-all"
                                                        title="Download"
                                                    >
                                                        <Download className="w-3.5 h-3.5" />
                                                    </button>
                                                    <button
                                                        onClick={(e) => { e.stopPropagation(); handleFileDelete(file); }}
                                                        className="p-1.5 bg-white border border-gray-200 rounded hover:bg-red-50 hover:border-red-200 text-gray-400 hover:text-red-600 transition-all"
                                                        title="Delete"
                                                    >
                                                        <Trash2 className="w-3.5 h-3.5" />
                                                    </button>
                                                </div>
                                                <span className="text-gray-400 text-[10px] font-bold uppercase tracking-widest hidden sm:block">
                                                    {file.is_folder ? 'Folder' : formatFileSize(file.size)}
                                                </span>
                                                <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-blue-500" />
                                            </div>
                                        </div>
                                    ))}
                                    {filteredFiles.length === 0 && !loader.loading && (
                                        <div className="p-20 text-center">
                                            <Folder className="w-16 h-16 text-gray-200 mx-auto mb-4" />
                                            <p className="text-gray-400">Empty directory</p>
                                        </div>
                                    )}
                                </>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </WithAdminBodyLayout>
    );
};

interface CreateFolderModalProps {
    handleCreateFolder: (name: string) => void;
    modal: any;
}

const CreateFolderModalContent = ({ handleCreateFolder, modal }: CreateFolderModalProps) => {
    const [folderName, setFolderName] = useState('');
    return (
        <div className="space-y-4">
            <div>
                <label className="block text-sm font-semibold text-gray-700 mb-2">Folder Name</label>
                <input
                    type="text"
                    value={folderName}
                    onChange={(e) => setFolderName(e.target.value)}
                    placeholder="Enter folder name"
                    className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent transition-all"
                    autoFocus
                    onKeyDown={(e) => {
                        if (e.key === 'Enter' && folderName.trim()) handleCreateFolder(folderName);
                        else if (e.key === 'Escape') modal.closeModal();
                    }}
                />
            </div>
            <div className="flex justify-end gap-2 pt-2">
                <button onClick={() => modal.closeModal()} className="px-4 py-2 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 font-semibold transition-all">Cancel</button>
                <button
                    onClick={() => handleCreateFolder(folderName)}
                    disabled={!folderName.trim()}
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 font-semibold shadow-sm transition-all"
                >
                    Create Folder
                </button>
            </div>
        </div>
    );
};

interface UploadModalProps {
    installId: number;
    currentPath: string;
    loader: any;
    modal: any;
    fileInputRef: React.RefObject<HTMLInputElement | null>;
    setUploading: (u: boolean) => void;
    uploading: boolean;
}

const UploadModalContent = ({ installId, currentPath, loader, modal, fileInputRef, setUploading, uploading }: UploadModalProps) => {
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
        <div className="space-y-4 max-h-[80vh] overflow-y-auto pr-2">
            <div className="flex gap-2 p-1 bg-gray-100 rounded-lg">
                <button
                    onClick={() => setUploadMode('regular')}
                    className={`flex-1 px-4 py-2 rounded-lg font-bold text-sm transition-all ${uploadMode === 'regular' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}`}
                >
                    Regular Upload
                </button>
                <button
                    onClick={() => setUploadMode('presigned')}
                    className={`flex-1 px-4 py-2 rounded-lg font-bold text-sm transition-all ${uploadMode === 'presigned' ? 'bg-white text-purple-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'}`}
                >
                    Presigned Upload
                </button>
            </div>

            {uploadMode === 'regular' ? (
                <div className="space-y-4">
                    <div className="bg-blue-50 border border-blue-100 rounded-xl p-4">
                        <h3 className="font-bold text-blue-900 text-sm mb-1 flex items-center gap-2">
                            <Upload className="w-4 h-4" />
                            Direct Upload
                        </h3>
                        <p className="text-xs text-blue-700 leading-relaxed font-medium">
                            Upload files directly to the current directory. Supports multi-selection.
                        </p>
                    </div>

                    <div className="border-2 border-dashed border-gray-200 rounded-xl p-10 text-center hover:border-blue-400 hover:bg-blue-50/30 transition-all group">
                        <Upload className="w-12 h-12 text-gray-300 mx-auto mb-4 group-hover:text-blue-400 transition-colors" />
                        <p className="text-sm font-bold text-gray-700 mb-1">Select files to upload</p>
                        <p className="text-xs font-semibold text-gray-400">Drag and drop also supported</p>
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
                                    alert('Upload failed: ' + error);
                                } finally {
                                    setUploading(false);
                                }
                            }}
                            className="hidden"
                        />
                        <button
                            onClick={() => fileInputRef.current?.click()}
                            disabled={uploading}
                            className="mt-6 px-6 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 font-bold shadow-md shadow-blue-500/20 transition-all"
                        >
                            {uploading ? 'Uploading...' : 'Choose Files'}
                        </button>
                    </div>
                </div>
            ) : (
                <div className="space-y-4">
                    <div className="bg-purple-50 border border-purple-100 rounded-xl p-4">
                        <h3 className="font-bold text-purple-900 text-sm mb-1 flex items-center gap-2">
                            <Key className="w-4 h-4" />
                            Secure Pre-signed Access
                        </h3>
                        <p className="text-xs text-purple-700 leading-relaxed font-medium">
                            Generate a 1-hour valid upload token for programmatic or third-party access.
                        </p>
                    </div>

                    <div className="space-y-4 bg-gray-50 rounded-xl p-5 border border-gray-200">
                        <div>
                            <label className="block text-xs font-bold text-gray-500 uppercase tracking-widest mb-2">Target Filename</label>
                            <input
                                type="text"
                                value={presignedFileName}
                                onChange={(e) => setPresignedFileName(e.target.value)}
                                placeholder="e.g., source-code.zip"
                                className="w-full px-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-purple-500 font-medium transition-all"
                            />
                        </div>
                        <button
                            onClick={handleGeneratePresignedURL}
                            disabled={!presignedFileName.trim()}
                            className="w-full py-2.5 bg-purple-600 text-white rounded-lg hover:bg-purple-700 disabled:opacity-50 font-bold shadow-md shadow-purple-500/20 transition-all"
                        >
                            Generate Access Token
                        </button>
                    </div>

                    {presignedData && (
                        <div className="space-y-4 border-t border-gray-200 pt-4">
                            <div className="flex items-center gap-2 text-green-600 text-xs font-bold">
                                <Check className="w-4 h-4" />
                                Token valid for 3600s
                            </div>
                            <div className="space-y-3">
                                <div>
                                    <label className="block text-[10px] font-bold text-gray-400 uppercase tracking-widest mb-1">Presigned Token</label>
                                    <div className="relative">
                                        <input type="text" value={presignedData.presigned_token} readOnly className="w-full px-3 py-2 bg-white border border-gray-200 rounded-lg font-mono text-[10px] text-gray-600 pr-16" />
                                        <button onClick={() => navigator.clipboard.writeText(presignedData.presigned_token)} className="absolute right-1 top-1 px-2 py-1 text-[10px] bg-gray-100 hover:bg-gray-200 rounded font-bold transition-all">Copy</button>
                                    </div>
                                </div>
                                <div className="bg-white border border-gray-200 rounded-xl p-4 space-y-4 shadow-sm">
                                    <div className="flex gap-2 overflow-x-auto pb-1">
                                        {['curl', 'javascript', 'python', 'browser'].map(t => (
                                            <button
                                                key={t}
                                                onClick={() => setExampleTab(t as any)}
                                                className={`px-3 py-1 text-[10px] font-bold rounded uppercase tracking-wider transition-all ${exampleTab === t ? 'bg-gray-900 text-white' : 'bg-gray-100 text-gray-500 hover:bg-gray-200'}`}
                                            >
                                                {t}
                                            </button>
                                        ))}
                                    </div>
                                    <div className="bg-gray-900 rounded-lg p-3 overflow-hidden">
                                        <pre className="text-[10px] text-green-400 font-mono leading-relaxed overflow-x-auto">
                                            {exampleTab === 'curl' && `curl -X POST \\\n  "${window.location.origin}${presignedData.upload_url}" \\\n  -F "file=@${presignedFileName}"`}
                                            {exampleTab === 'javascript' && `const formData = new FormData();\nformData.append('file', file);\n\nfetch('${window.location.origin}${presignedData.upload_url}', {\n  method: 'POST',\n  body: formData\n});`}
                                            {exampleTab === 'python' && `import requests\nfiles = {'file': open('${presignedFileName}', 'rb')}\nrequests.post('${window.location.origin}${presignedData.upload_url}', files=files)`}
                                            {exampleTab === 'browser' && `${window.location.origin}/zz/pages/presigned/file?presigned-key=${presignedData.presigned_token}`}
                                        </pre>
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            )}

            <div className="flex justify-end pt-2">
                <button onClick={() => modal.closeModal()} className="px-4 py-2 text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 font-bold transition-all">Close</button>
            </div>
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

interface FileEditorProps {
    installId: number;
    file: SpaceFile;
    initialContent: string;
    currentPath: string;
    onSave: () => void;
    onCancel: () => void;
}

const FileEditor = ({ installId, file, initialContent, currentPath, onSave, onCancel }: FileEditorProps) => {
    const [content, setContent] = useState(initialContent);
    const [saving, setSaving] = useState(false);
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        // Focus the textarea when component mounts
        if (textareaRef.current) {
            textareaRef.current.focus();
        }
    }, []);

    const handleSave = async () => {
        setSaving(true);
        try {
            await updateSpaceFileContent(installId, file.id, content, file.name, currentPath);
            onSave();
        } catch (error) {
            console.error('Failed to save file:', error);
            alert('Failed to save file. Please try again.');
        } finally {
            setSaving(false);
        }
    };

    return (
        <div className="flex flex-col h-full max-h-[80vh]">
            <div className="flex-1 mb-4">
                <textarea
                    ref={textareaRef}
                    value={content}
                    onChange={(e) => setContent(e.target.value)}
                    className="w-full h-full p-4 border border-gray-300 rounded-lg font-mono text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                    style={{ minHeight: '400px' }}
                    spellCheck={false}
                />
            </div>
            <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                <div className="text-sm text-gray-500">
                    {content.length} characters
                </div>
                <div className="flex gap-2">
                    <button
                        onClick={onCancel}
                        disabled={saving}
                        className="px-4 py-2 text-gray-700 bg-gray-100 rounded-lg hover:bg-gray-200 disabled:opacity-50"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={handleSave}
                        disabled={saving || content === initialContent}
                        className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        {saving ? 'Saving...' : 'Save'}
                    </button>
                </div>
            </div>
        </div>
    );
};
