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
    Edit
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { listPackageFiles, PackageFile, deletePackageFile, downloadPackageFile, uploadPackageFile, updatePackageFileContent } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { useGApp } from '@/hooks';
import BigSearchBar from '@/contain/compo/BigSearchBar';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const packageVersionId = searchParams.get('package_version_id');
    const gapp = useGApp();

    if (!packageVersionId) {
        return (
            <WithAdminBodyLayout Icon={Folder} name="Package Files" description="Select a package to view files">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <Folder className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                        <p className="text-gray-500">No package selected</p>
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

    return <FileManager packageId={parseInt(packageVersionId)} />;
}

interface FileManagerProps {
    packageId: number;
}

const FileManager = ({ packageId }: FileManagerProps) => {
    const [currentPath, setCurrentPath] = useState('');
    const [searchTerm, setSearchTerm] = useState('');
    const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
    const [uploading, setUploading] = useState(false);
    const fileInputRef = useRef<HTMLInputElement>(null);
    const gapp = useGApp();

    const loader = useSimpleDataLoader<PackageFile[]>({
        loader: () => listPackageFiles(packageId, currentPath),
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

    const handleFolderClick = (folder: PackageFile) => {
        const newPath = currentPath ? `${currentPath}/${folder.name}` : folder.name;
        setCurrentPath(newPath);
    };

    const handleBackClick = () => {
        const pathParts = currentPath.split('/');
        pathParts.pop();
        setCurrentPath(pathParts.join('/'));
    };

    const handleFileDownload = async (file: PackageFile) => {
        try {
            const response = await downloadPackageFile(packageId, file.id);
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

    const handleFileDelete = async (file: PackageFile) => {
        if (!confirm(`Are you sure you want to delete "${file.name}"?`)) return;

        try {
            await deletePackageFile(packageId, file.id);
            loader.reload();
        } catch (error) {
            console.error('Delete failed:', error);
        }
    };

    const handleFileEdit = async (file: PackageFile) => {
        try {
            // Download file content
            const response = await downloadPackageFile(packageId, file.id);
            
            // Ensure we're getting a blob response
            if (!(response.data instanceof Blob)) {
                throw new Error('Expected blob response but got something else');
            }
            
            let content = await response.data.text();

            // Open modal with editor
            gapp.modal.openModal({
                title: `Edit ${file.name}`,
                content: (
                    <FileEditor
                        packageId={packageId}
                        file={file}
                        initialContent={content}
                        currentPath={currentPath}
                        onSave={() => {
                            loader.reload();
                            gapp.modal.closeModal();
                        }}
                        onCancel={() => gapp.modal.closeModal()}
                    />
                ),
                size: 'xl'
            });
        } catch (error) {
            console.error('Failed to load file for editing:', error);
            alert('Failed to load file for editing. The file might be too large, not a text file, or corrupted.');
        }
    };

    const handleFileUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
        const files = event.target.files;
        if (!files || files.length === 0) return;

        setUploading(true);
        try {
            for (const file of Array.from(files)) {
                await uploadPackageFile(packageId, file, currentPath);
            }
            loader.reload();
        } catch (error) {
            console.error('Upload failed:', error);
        } finally {
            setUploading(false);
            if (fileInputRef.current) {
                fileInputRef.current.value = '';
            }
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Folder}
            name="Package Files"
            description={`Managing files for package`}
            rightContent={
                <div className="flex items-center gap-2">
                    <button
                        onClick={() => fileInputRef.current?.click()}
                        disabled={uploading}
                        className="flex items-center gap-2 px-3 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50"
                    >
                        <Upload className="w-4 h-4" />
                        {uploading ? 'Uploading...' : 'Upload'}
                    </button>
                    <input
                        ref={fileInputRef}
                        type="file"
                        multiple
                        onChange={handleFileUpload}
                        className="hidden"
                    />
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
                                onEdit={() => { }}
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
                                onEdit={() => handleFileEdit(file)}
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
    file: PackageFile;
    viewMode: 'list' | 'grid';
    onDoubleClick: () => void;
    onDownload: () => void;
    onEdit: () => void;
    onDelete: () => void;
}

const FileItem = ({ file, viewMode, onDoubleClick, onDownload, onEdit, onDelete }: FileItemProps) => {
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
                                onEdit();
                            }}
                            className="p-1 bg-white rounded shadow-sm hover:bg-gray-100"
                            title="Edit file"
                        >
                            <Edit className="w-3 h-3" />
                        </button>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onDownload();
                            }}
                            className="p-1 bg-white rounded shadow-sm hover:bg-gray-100"
                            title="Download file"
                        >
                            <Download className="w-3 h-3" />
                        </button>
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onDelete();
                            }}
                            className="p-1 bg-white rounded shadow-sm hover:bg-gray-100 text-red-500"
                            title="Delete file"
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
                            onEdit();
                        }}
                        className="p-1 hover:bg-gray-200 rounded"
                        title="Edit file"
                    >
                        <Edit className="w-4 h-4" />
                    </button>
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            onDownload();
                        }}
                        className="p-1 hover:bg-gray-200 rounded"
                        title="Download file"
                    >
                        <Download className="w-4 h-4" />
                    </button>
                    <button
                        onClick={(e) => {
                            e.stopPropagation();
                            onDelete();
                        }}
                        className="p-1 hover:bg-gray-200 rounded text-red-500"
                        title="Delete file"
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

interface FileEditorProps {
    packageId: number;
    file: PackageFile;
    initialContent: string;
    currentPath: string;
    onSave: () => void;
    onCancel: () => void;
}

const FileEditor = ({ packageId, file, initialContent, currentPath, onSave, onCancel }: FileEditorProps) => {
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
    
            await updatePackageFileContent(packageId, file.id, content, file.name, currentPath);
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
