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
    Edit,
    Copy,
    RefreshCw,
    Upload,
    ChevronDown
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { useGApp } from '@/hooks';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { getInstalledPackageInfo, listPackageFiles, downloadPackageFile, updatePackageFileContent, PackageVersion, PackageFile, InstalledPackageInfo, generatePackageDevToken, listPackageAvailableVersions, upgradePackageFromRepo, upgradePackageZipDirectly, AvailableVersionsResponse, UpgradePackageResult } from '@/lib';

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
    const [generatingToken, setGeneratingToken] = useState(false);
    const [dropdownOpen, setDropdownOpen] = useState(false);
    const gapp = useGApp();
    const { modal } = gapp;
    const dropdownRef = React.useRef<HTMLDivElement>(null);

    useEffect(() => {
        const close = (e: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
                setDropdownOpen(false);
            }
        };
        document.addEventListener('click', close);
        return () => document.removeEventListener('click', close);
    }, []);

    const loader = useSimpleDataLoader<InstalledPackageInfo>({
        loader: () => getInstalledPackageInfo(packageId),
        ready: gapp.isInitialized,
    });

    const packageVersions = loader.data?.package_versions || [];
    const packageData = loader.data?.installed_package;

    const handleGenerateDevToken = async () => {
        modal.openModal({
            title: "Development Token",
            content: <DevTokenModalContent packageId={packageId} modal={modal} />,
            size: "sm"
        });
    };

    const handleCheckForUpdate = () => {
        modal.openModal({
            title: "Check for update",
            content: (
                <CheckForUpdateModalContent
                    packageId={packageId}
                    packageName={packageData?.name}
                    modal={modal}
                    onUpdated={() => {
                        loader.reload();
                        modal.closeModal();
                    }}
                />
            ),
            size: "md"
        });
    };

    const handleUpgradeFromZip = () => {
        modal.openModal({
            title: "Upgrade from ZIP",
            content: (
                <UpgradeFromZipModalContent
                    packageId={packageId}
                    modal={modal}
                    onUpdated={() => {
                        loader.reload();
                        modal.closeModal();
                    }}
                />
            ),
            size: "sm"
        });
    };

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
                rightContent={
                    <div className="flex items-center gap-2">
                        <button
                            onClick={handleCheckForUpdate}
                            className="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 shadow-sm font-semibold transition-all"
                        >
                            <RefreshCw className="w-4 h-4 text-green-600" />
                            Update
                        </button>
                        <div className="relative" ref={dropdownRef}>
                            <button
                                type="button"
                                onClick={() => setDropdownOpen((o) => !o)}
                                className="flex items-center gap-2 px-3 py-2 bg-white border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-50 shadow-sm font-semibold transition-all"
                            >
                                More
                                <ChevronDown className={`w-4 h-4 transition-transform ${dropdownOpen ? 'rotate-180' : ''}`} />
                            </button>
                            {dropdownOpen && (
                                <div className="absolute right-0 top-full mt-1 py-1 min-w-[180px] bg-white border border-gray-200 rounded-lg shadow-lg z-10">
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setDropdownOpen(false);
                                            handleUpgradeFromZip();
                                        }}
                                        className="w-full flex items-center gap-2 px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
                                    >
                                        <Upload className="w-4 h-4 text-amber-600 shrink-0" />
                                        Upgrade from ZIP
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setDropdownOpen(false);
                                            handleGenerateDevToken();
                                        }}
                                        className="w-full flex items-center gap-2 px-4 py-2 text-left text-sm text-gray-700 hover:bg-gray-50"
                                    >
                                        <Key className="w-4 h-4 text-blue-500 shrink-0" />
                                        Dev Token
                                    </button>
                                </div>
                            )}
                        </div>
                    </div>
                }
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

interface UpgradeFromZipModalProps {
    packageId: number;
    modal: any;
    onUpdated: () => void;
}

const UpgradeFromZipModalContent = ({ packageId, modal, onUpdated }: UpgradeFromZipModalProps) => {
    const router = useRouter();
    const [selectedFile, setSelectedFile] = useState<File | null>(null);
    const [uploading, setUploading] = useState(false);
    const [upgradeResult, setUpgradeResult] = useState<UpgradePackageResult | null>(null);
    const fileInputRef = React.useRef<HTMLInputElement>(null);

    const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const f = e.target.files?.[0];
        if (f) setSelectedFile(f);
    };

    const handleUpgrade = async () => {
        if (!selectedFile) return;
        setUploading(true);
        setUpgradeResult(null);
        try {
            const buffer = await selectedFile.arrayBuffer();
            const res = await upgradePackageZipDirectly(packageId, buffer);
            const result = res.data;
            if (result?.update_page) {
                setUpgradeResult(result);
            } else {
                onUpdated();
                modal.closeModal();
            }
        } catch (err: any) {
            alert(err?.response?.data?.message || err?.message || 'Upgrade failed');
        } finally {
            setUploading(false);
        }
    };

    const handleConfigure = () => {
        if (!upgradeResult?.update_page) return;
        const fragment = new URLSearchParams();
        fragment.set('nskey', upgradeResult.key_space);
        fragment.set('space_id', upgradeResult.root_space_id.toString());
        fragment.set('load_page', upgradeResult.update_page);
        router.push(`/portal/admin/exec?${fragment.toString()}`);
        onUpdated();
        modal.closeModal();
    };

    if (upgradeResult?.update_page) {
        return (
            <div className="space-y-4">
                <p className="text-sm text-gray-600">Upgrade completed. You can configure the package or close.</p>
                <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                    <button onClick={() => { onUpdated(); modal.closeModal(); }} className="px-4 py-2 text-gray-600 hover:text-gray-800 font-medium">
                        Close
                    </button>
                    <button onClick={handleConfigure} className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 font-medium">
                        Configure
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-600">
                Upload a package ZIP file to upgrade this installation. The ZIP must contain a valid potato package.
            </p>
            <div className="flex items-center gap-3">
                <input
                    ref={fileInputRef}
                    type="file"
                    accept=".zip"
                    onChange={handleFileChange}
                    className="hidden"
                />
                <button
                    type="button"
                    onClick={() => fileInputRef.current?.click()}
                    className="flex items-center gap-2 px-4 py-2 bg-gray-100 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-200 font-medium transition-all"
                >
                    <Upload className="w-4 h-4" />
                    Choose ZIP file
                </button>
                {selectedFile && (
                    <span className="text-sm text-gray-600 truncate max-w-[200px]" title={selectedFile.name}>
                        {selectedFile.name}
                    </span>
                )}
            </div>
            <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                <button onClick={() => modal.closeModal()} className="px-4 py-2 text-gray-600 hover:text-gray-800 font-medium">
                    Cancel
                </button>
                <button
                    onClick={handleUpgrade}
                    disabled={!selectedFile || uploading}
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                    {uploading ? 'Upgrading…' : 'Upgrade'}
                </button>
            </div>
        </div>
    );
};

interface CheckForUpdateModalProps {
    packageId: number;
    packageName?: string;
    modal: any;
    onUpdated: () => void;
}

const CheckForUpdateModalContent = ({ packageId, packageName, modal, onUpdated }: CheckForUpdateModalProps) => {
    const router = useRouter();
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [data, setData] = useState<AvailableVersionsResponse | null>(null);
    const [selectedVersion, setSelectedVersion] = useState<string | null>(null);
    const [updating, setUpdating] = useState(false);
    const [upgradeResult, setUpgradeResult] = useState<UpgradePackageResult | null>(null);

    useEffect(() => {
        let cancelled = false;
        setLoading(true);
        setError(null);
        listPackageAvailableVersions(packageId)
            .then((res) => {
                if (!cancelled) {
                    setData(res.data);
                    setSelectedVersion(null);
                }
            })
            .catch((err) => {
                if (!cancelled) {
                    setError(err?.response?.data?.message || err?.message || 'Failed to load versions');
                }
            })
            .finally(() => {
                if (!cancelled) setLoading(false);
            });
        return () => { cancelled = true; };
    }, [packageId]);

    const handleUpdate = async () => {
        if (!data || !selectedVersion) return;
        setUpdating(true);
        setUpgradeResult(null);
        try {
            const res = await upgradePackageFromRepo(packageId, {
                repo_slug: data.repo_slug,
                name: data.name,
                version: selectedVersion,
            });
            const result = res.data;
            if (result?.update_page) {
                setUpgradeResult(result);
            } else {
                onUpdated();
                modal.closeModal();
            }
        } catch (err: any) {
            alert(err?.response?.data?.message || err?.message || 'Update failed');
        } finally {
            setUpdating(false);
        }
    };

    const handleConfigure = () => {
        if (!upgradeResult?.update_page) return;
        const fragment = new URLSearchParams();
        fragment.set('nskey', upgradeResult.key_space);
        fragment.set('space_id', upgradeResult.root_space_id.toString());
        fragment.set('load_page', upgradeResult.update_page);
        router.push(`/portal/admin/exec?${fragment.toString()}`);
        onUpdated();
        modal.closeModal();
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center py-12">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500" />
            </div>
        );
    }
    if (error) {
        return (
            <div className="space-y-4">
                <p className="text-sm text-red-600">{error}</p>
                <div className="flex justify-end">
                    <button onClick={() => modal.closeModal()} className="px-4 py-2 text-gray-600 hover:text-gray-800 font-medium">Close</button>
                </div>
            </div>
        );
    }
    if (!data?.versions?.length) {
        return (
            <div className="space-y-4">
                <p className="text-sm text-gray-600">
                    {data?.repo_slug
                        ? `No versions found in repo for ${packageName || data.name}.`
                        : 'This package was not installed from a repo. Use a ZIP upload to update.'}
                </p>
                <div className="flex justify-end">
                    <button onClick={() => modal.closeModal()} className="px-4 py-2 text-gray-600 hover:text-gray-800 font-medium">Close</button>
                </div>
            </div>
        );
    }

    if (upgradeResult?.update_page) {
        return (
            <div className="space-y-4">
                <p className="text-sm text-gray-600">Update completed. You can configure the package or close.</p>
                <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                    <button onClick={() => { onUpdated(); modal.closeModal(); }} className="px-4 py-2 text-gray-600 hover:text-gray-800 font-medium">
                        Close
                    </button>
                    <button onClick={handleConfigure} className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 font-medium">
                        Configure
                    </button>
                </div>
            </div>
        );
    }

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-600">
                Select a version to install. Current: <strong>{data.current_version || '—'}</strong>
            </p>
            <div className="max-h-64 overflow-y-auto border border-gray-200 rounded-lg divide-y divide-gray-100">
                {data.versions.map((v) => (
                    <label
                        key={v}
                        className={`flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-gray-50 ${selectedVersion === v ? 'bg-blue-50' : ''}`}
                    >
                        <input
                            type="radio"
                            name="version"
                            checked={selectedVersion === v}
                            onChange={() => setSelectedVersion(v)}
                            className="text-blue-600"
                        />
                        <span className="font-medium">v{v}</span>
                        {v === data.current_version && (
                            <span className="text-xs bg-gray-200 text-gray-700 px-2 py-0.5 rounded">Current</span>
                        )}
                    </label>
                ))}
            </div>
            <div className="flex justify-end gap-2 pt-2 border-t border-gray-100">
                <button onClick={() => modal.closeModal()} className="px-4 py-2 text-gray-600 hover:text-gray-800 font-medium">
                    Cancel
                </button>
                <button
                    onClick={handleUpdate}
                    disabled={!selectedVersion || updating}
                    className="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-medium"
                >
                    {updating ? 'Updating…' : 'Update'}
                </button>
            </div>
        </div>
    );
};

interface DevTokenModalProps {
    packageId: number;
    modal: any;
}

const DevTokenModalContent = ({ packageId, modal }: DevTokenModalProps) => {
    const [devToken, setDevToken] = useState<string>('');
    const [copied, setCopied] = useState(false);
    const [generatingToken, setGeneratingToken] = useState(false);

    const handleGenerateDevToken = async () => {
        try {
            setGeneratingToken(true);
            const response = await generatePackageDevToken(packageId);
            setDevToken(response.data.token);
        } catch (error) {
            console.error('Failed to generate dev token:', error);
            alert('Failed to generate dev token. Please try again.');
        } finally {
            setGeneratingToken(false);
        }
    };

    const handleCopyToken = () => {
        if (devToken) {
            navigator.clipboard.writeText(devToken);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        }
    };

    return (
        <div className="space-y-4">
            <p className="text-sm text-gray-600">
                Generate a development token for CLI package push operations.
                This token allows you to update this package from the command line.
            </p>

            {!devToken ? (
                <button
                    onClick={handleGenerateDevToken}
                    disabled={generatingToken}
                    className="w-full flex items-center justify-center gap-2 px-4 py-3 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed font-bold transition-all shadow-sm"
                >
                    <Key className="w-5 h-5" />
                    {generatingToken ? 'Generating...' : 'Generate New Dev Token'}
                </button>
            ) : (
                <div className="space-y-4">
                    <div className="relative group">
                        <div className="w-full bg-gray-50 border border-gray-200 rounded-lg p-4 font-mono text-xs break-all pr-12 min-h-[80px] flex items-center">
                            {devToken}
                        </div>
                        <button
                            onClick={handleCopyToken}
                            className="absolute top-2 right-2 p-2 bg-white border border-gray-200 rounded-md hover:bg-gray-50 transition-all shadow-sm"
                            title="Copy to clipboard"
                        >
                            {copied ? (
                                <Check className="w-4 h-4 text-green-500" />
                            ) : (
                                <Copy className="w-4 h-4 text-gray-400" />
                            )}
                        </button>
                    </div>
                    {copied && (
                        <p className="text-xs text-green-600 font-bold flex items-center gap-1 justify-center">
                            <Check className="w-3 h-3" />
                            Token copied to clipboard!
                        </p>
                    )}
                    <button
                        onClick={handleGenerateDevToken}
                        disabled={generatingToken}
                        className="w-full text-sm text-blue-500 hover:text-blue-600 font-bold py-2 border border-blue-100 rounded-lg hover:bg-blue-50 transition-all"
                    >
                        Generate Different Token
                    </button>
                </div>
            )}

            <div className="pt-4 border-t border-gray-100 flex justify-end">
                <button
                    onClick={() => modal.closeModal()}
                    className="px-4 py-2 text-gray-500 hover:text-gray-700 font-bold"
                >
                    Close
                </button>
            </div>
        </div>
    );
};
