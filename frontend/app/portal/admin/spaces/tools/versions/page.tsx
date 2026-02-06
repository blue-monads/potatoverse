"use client";
import React, { useState, useEffect } from 'react';
import {
    History,
    Tag,
    ChevronRight,
    Search,
    Package,
    ArrowLeft,
    Folder,
    File,
    Download,
    ChevronLeft,
    Box,
    Globe,
    ExternalLink,
    Code2,
    Scale,
    Calendar,
    CloudLightning,
    Zap,
    Users,
    Key,
    Info,
    Check,
    Clock,
    Edit
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { useGApp } from '@/hooks';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { getInstalledPackageInfo, listPackageFiles, downloadPackageFile, updatePackageFileContent, PackageVersion, PackageFile, InstalledPackageInfo } from '@/lib';

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const gapp = useGApp();

    if (!installId) {
        return (
            <WithAdminBodyLayout Icon={History} name="Package Versions" description="Select a package to view versions" variant="none">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <History className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                        <p className="text-gray-500">No package selected</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    return <VersionsManager packageId={parseInt(installId)} />;
}

interface VersionsManagerProps {
    packageId: number;
}

const VersionsManager = ({ packageId }: VersionsManagerProps) => {
    const [selectedVersionId, setSelectedVersionId] = useState<number | null>(null);
    const [searchTerm, setSearchTerm] = useState('');
    const gapp = useGApp();

    const loader = useSimpleDataLoader<InstalledPackageInfo>({
        loader: () => getInstalledPackageInfo(packageId),
        ready: gapp.isInitialized,
    });

    const packageVersions = loader.data?.package_versions || [];
    const packageData = loader.data?.installed_package;

    if (loader.loading) {
        return (
            <WithAdminBodyLayout Icon={History} name="Versions" description="Loading versions..." variant="none">
                <div className="flex items-center justify-center h-64">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    if (selectedVersionId) {
        const version = packageVersions.find(v => v.id === selectedVersionId);
        return (
            <VersionFileBrowser
                packageId={packageId}
                version={version!}
                onBack={() => setSelectedVersionId(null)}
            />
        );
    }

    return (
        <WithAdminBodyLayout
            Icon={History}
            name="Package Versions"
            description={`Versions for ${packageData?.name || 'Package'}`}
            variant="none"
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
                placeholder="Search versions..."
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
                <div className="grid grid-cols-1 gap-4">
                    {packageVersions.length === 0 ? (
                        <div className="text-center py-20 bg-white rounded-xl border border-dashed border-gray-300">
                            <Tag className="w-12 h-12 text-gray-400 mx-auto mb-4" />
                            <p className="text-gray-500">No versions found for this package</p>
                        </div>
                    ) : (
                        packageVersions
                            .filter(v => v.version.toLowerCase().includes(searchTerm.toLowerCase()) || v.name.toLowerCase().includes(searchTerm.toLowerCase()))
                            .map((version) => (
                                <div
                                    key={version.id}
                                    onClick={() => setSelectedVersionId(version.id)}
                                    className="group flex items-center justify-between p-5 bg-white border border-gray-200 rounded-xl hover:border-blue-400 hover:shadow-md transition-all cursor-pointer"
                                >
                                    <div className="flex items-center gap-4">
                                        <div className="w-12 h-12 bg-blue-50 rounded-lg flex items-center justify-center group-hover:bg-blue-600 transition-colors">
                                            <Tag className="w-6 h-6 text-blue-600 group-hover:text-white" />
                                        </div>
                                        <div>
                                            <h3 className="text-lg font-bold text-gray-900 flex items-center gap-2">
                                                {version.name}
                                                {version.id === packageData?.active_install_id && (
                                                    <span className="bg-blue-100 text-blue-700 text-[10px] font-bold px-2 py-0.5 rounded-full uppercase">
                                                        Active
                                                    </span>
                                                )}
                                            </h3>
                                            <div className="flex items-center gap-3 mt-1">
                                                <span className="text-sm font-semibold text-blue-600">v{version.version}</span>
                                                <span className="text-gray-400 text-xs">•</span>
                                                <span className="text-gray-500 text-xs flex items-center gap-1">
                                                    <Clock className="w-3 h-3" />
                                                    ID: #{version.id}
                                                </span>
                                                {version.format_version && (
                                                    <>
                                                        <span className="text-gray-400 text-xs">•</span>
                                                        <span className="text-gray-500 text-xs">Format: {version.format_version}</span>
                                                    </>
                                                )}
                                            </div>
                                        </div>
                                    </div>
                                    <div className="flex items-center gap-3">
                                        <div className="text-right hidden md:block">
                                            <p className="text-xs text-gray-400 uppercase font-bold tracking-wider">Source Files</p>
                                            <p className="text-sm font-bold text-gray-600">Browse code &rarr;</p>
                                        </div>
                                        <ChevronRight className="w-5 h-5 text-gray-300 group-hover:text-blue-500 transition-colors" />
                                    </div>
                                </div>
                            ))
                    )}
                </div>
            </div>
        </WithAdminBodyLayout>
    );
};

interface VersionFileBrowserProps {
    packageId: number;
    version: PackageVersion;
    onBack: () => void;
}

const VersionFileBrowser = ({ packageId, version, onBack }: VersionFileBrowserProps) => {
    const [currentPath, setCurrentPath] = useState('');
    const gapp = useGApp();

    const loader = useSimpleDataLoader<PackageFile[]>({
        loader: () => listPackageFiles(version.id, currentPath),
        ready: gapp.isInitialized,
        dependencies: [currentPath],
    });

    const handleBackClick = () => {
        const pathParts = currentPath.split('/');
        pathParts.pop();
        setCurrentPath(pathParts.join('/'));
    };

    const handleFileDownload = async (file: PackageFile) => {
        try {
            const response = await downloadPackageFile(version.id, file.id);
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

    const handleFileEdit = async (file: PackageFile) => {
        try {
            const response = await downloadPackageFile(version.id, file.id);
            if (!(response.data instanceof Blob)) {
                throw new Error('Expected blob response but got something else');
            }
            let content = await response.data.text();

            gapp.modal.openModal({
                title: `Edit ${file.name}`,
                content: (
                    <FileEditor
                        versionId={version.id}
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
            alert('Failed to load file for editing. The file might be too large or not a text file.');
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={History}
            name="Version Files"
            description={`Browsing files for ${version.name} (v${version.version})`}
            variant="none"
        >
            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
                {/* Reference UI Title Header */}
                <div className="flex items-center justify-between mb-6">
                    <button
                        onClick={onBack}
                        className="flex items-center gap-2 text-gray-600 hover:text-blue-600 font-bold transition-colors"
                    >
                        <ChevronLeft className="w-5 h-5" />
                        Back to Versions
                    </button>
                    <div className="flex items-center gap-2 text-sm text-gray-400">
                        <Tag className="w-4 h-4" />
                        <span>Fixed Version {version.version}</span>
                    </div>
                </div>

                {/* Light Themed File Browser */}
                <div className="bg-white rounded-xl overflow-hidden shadow-sm border border-gray-200">
                    {/* Header */}
                    <div className="bg-gray-50 px-6 py-4 flex items-center gap-3 border-b border-gray-200">
                        <div className="w-8 h-8 bg-blue-50 rounded flex items-center justify-center">
                            <Folder className="w-4 h-4 text-blue-500" />
                        </div>
                        <h2 className="text-gray-900 font-bold tracking-wide">
                            {currentPath ? `Package root / ${currentPath.split('/').join(' / ')}` : 'Package root'}
                        </h2>
                    </div>

                    {/* File List */}
                    <div className="divide-y divide-gray-100">
                        {loader.loading ? (
                            <div className="p-20 text-center">
                                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                                <p className="text-gray-500 text-sm">Loading source files...</p>
                            </div>
                        ) : loader.data?.length === 0 ? (
                            <div className="p-20 text-center">
                                <Folder className="w-16 h-16 text-gray-200 mx-auto mb-4" />
                                <p className="text-gray-400">Empty directory</p>
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
                                {loader.data?.map(file => (
                                    <div
                                        key={file.id}
                                        className="px-6 py-4 hover:bg-gray-50 cursor-pointer group flex items-center transition-colors"
                                        onClick={() => file.is_folder ? setCurrentPath(currentPath ? `${currentPath}/${file.name}` : file.name) : handleFileDownload(file)}
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
                                            <div className="hidden group-hover:flex items-center gap-2 mr-2">
                                                {!file.is_folder && (
                                                    <button
                                                        onClick={(e) => {
                                                            e.stopPropagation();
                                                            handleFileEdit(file);
                                                        }}
                                                        className="p-1.5 bg-white border border-gray-200 rounded hover:bg-blue-50 hover:border-blue-200 text-gray-500 hover:text-blue-600 transition-all"
                                                        title="Edit file"
                                                    >
                                                        <Edit className="w-3.5 h-3.5" />
                                                    </button>
                                                )}
                                                <button
                                                    onClick={(e) => {
                                                        e.stopPropagation();
                                                        handleFileDownload(file);
                                                    }}
                                                    className="p-1.5 bg-white border border-gray-200 rounded hover:bg-blue-50 hover:border-blue-200 text-gray-500 hover:text-blue-600 transition-all"
                                                    title="Download"
                                                >
                                                    <Download className="w-3.5 h-3.5" />
                                                </button>
                                            </div>
                                            <span className="text-gray-400 text-[10px] font-bold uppercase tracking-widest hidden sm:block">
                                                {file.is_folder ? 'Folder' : formatFileSize(file.size)}
                                            </span>
                                            <ChevronRight className="w-4 h-4 text-gray-300 group-hover:text-blue-500" />
                                        </div>
                                    </div>
                                ))}
                            </>
                        )}
                    </div>
                </div>
            </div>
        </WithAdminBodyLayout>
    );
};

interface FileEditorProps {
    versionId: number;
    file: PackageFile;
    initialContent: string;
    currentPath: string;
    onSave: () => void;
    onCancel: () => void;
}

const FileEditor = ({ versionId, file, initialContent, currentPath, onSave, onCancel }: FileEditorProps) => {
    const [content, setContent] = useState(initialContent);
    const [saving, setSaving] = useState(false);
    const textareaRef = React.useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        if (textareaRef.current) {
            textareaRef.current.focus();
        }
    }, []);

    const handleSave = async () => {
        setSaving(true);
        try {
            await updatePackageFileContent(versionId, file.id, content, file.name, currentPath);
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
                    className="w-full h-full p-4 border border-gray-300 rounded-lg font-mono text-sm resize-none focus:outline-none focus:ring-2 focus:ring-blue-500 bg-gray-50"
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

const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
};
