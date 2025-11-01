"use client";
import React, { useState, useEffect } from 'react';
import {
    Package as PackageIcon,
    Box,
    Code2,
    Globe,
    Mail,
    Github,
    Scale,
    Tag,
    Calendar,
    Copy,
    Check,
    Key,
    ExternalLink,
    Info
} from 'lucide-react';
import { useRouter, useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { useGApp } from '@/hooks';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { getInstalledPackageInfo, InstalledPackageInfo, generatePackageDevToken, Space, PackageVersion } from '@/lib';
import { staticGradients } from '@/app/utils';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');
    const namespaceKey = searchParams.get('namespace_key');
    const packageVersionId = searchParams.get('package_version_id');



    const gapp = useGApp();

    if (!installId) {
        return (
            <WithAdminBodyLayout Icon={PackageIcon} name="Package About" description="Select a package to view details">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <PackageIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
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

    return <PackageAbout packageId={parseInt(installId)} spaceId={spaceId ? parseInt(spaceId) : undefined} />;
}

interface PackageAboutProps {
    packageId: number;
    spaceId?: number;
}

const PackageAbout = ({ packageId, spaceId }: PackageAboutProps) => {
    const gapp = useGApp();
    const [devToken, setDevToken] = useState<string>('');
    const [copied, setCopied] = useState(false);
    const [generatingToken, setGeneratingToken] = useState(false);

    const loader = useSimpleDataLoader<InstalledPackageInfo>({
        loader: () => getInstalledPackageInfo(packageId),
        ready: gapp.isInitialized,
    });

    const packageData = loader.data?.installed_package;
    const packageVersions = loader.data?.package_versions || [];
    const packageSpaces = loader.data?.spaces || [];
    const activeVersion = packageData ? packageVersions.find(v => v.id === packageData.active_install_id) : null;

    const handleGenerateDevToken = async () => {
        try {
            setGeneratingToken(true);
            const response = await generatePackageDevToken(packageId);
            setDevToken(response.data.token);
        } catch (error) {
            console.error('Failed to generate dev token:', error);
            gapp.modal.openModal({
                title: "Error",
                content: (
                    <div className="text-red-600">
                        Failed to generate dev token. Please try again.
                    </div>
                ),
                size: "sm"
            });
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

    if (loader.loading) {
        return (
            <WithAdminBodyLayout Icon={PackageIcon} name="Package About" description="Loading package details...">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                        <p className="text-gray-500">Loading package information...</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    if (!packageData) {
        return (
            <WithAdminBodyLayout Icon={PackageIcon} name="Package About" description="Package not found">
                <div className="flex items-center justify-center h-64">
                    <div className="text-center">
                        <PackageIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                        <p className="text-gray-500">Package not found</p>
                    </div>
                </div>
            </WithAdminBodyLayout>
        );
    }

    return (
        <WithAdminBodyLayout 
            Icon={PackageIcon} 
            name="Package About" 
            description={`Details for ${packageData.name}`}
        >
            <div className="max-w-5xl mx-auto w-full px-4 py-6">
            <div className="card p-6 flex flex-col gap-6">
                {/* Package Header */}
                <div className="bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg p-6 text-white">
                    <div className="flex items-start justify-between">
                        <div>
                            <div className="flex items-center gap-3 mb-2">
                                <PackageIcon className="w-8 h-8" />
                                <h1 className="text-3xl font-bold">{packageData.name}</h1>
                            </div>
                            {activeVersion && (
                                <p className="text-blue-100 mb-4">{activeVersion.info || 'No description available'}</p>
                            )}
                            <div className="flex items-center gap-4">
                                {activeVersion && (
                                    <span className="bg-white/20 px-3 py-1 rounded-full text-sm">
                                        v{activeVersion.version}
                                    </span>
                                )}
                                <span className="bg-white/20 px-3 py-1 rounded-full text-sm">
                                    ID: {packageData.id}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Package Details Grid */}
                {activeVersion && (
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        {/* Left Column */}
                        <div className="space-y-4">
                            <DetailCard
                                icon={<Code2 className="w-5 h-5" />}
                                label="Slug"
                                value={activeVersion.slug || 'N/A'}
                            />
                            <DetailCard
                                icon={<Tag className="w-5 h-5" />}
                                label="Tags"
                                value={activeVersion.tags || 'No tags'}
                            />
                            <DetailCard
                                icon={<Scale className="w-5 h-5" />}
                                label="License"
                                value={activeVersion.license || 'N/A'}
                            />
                        </div>

                        {/* Right Column */}
                        <div className="space-y-4">
                            <DetailCard
                                icon={<Calendar className="w-5 h-5" />}
                                label="Format Version"
                                value={activeVersion.format_version || 'N/A'}
                            />
                            <DetailCard
                                icon={<Info className="w-5 h-5" />}
                                label="Description"
                                value={activeVersion.info || 'No description'}
                            />
                            {activeVersion.author_name && (
                                <DetailCard
                                    icon={<Mail className="w-5 h-5" />}
                                    label="Author"
                                    value={`${activeVersion.author_name}${activeVersion.author_email ? ` (${activeVersion.author_email})` : ''}`}
                                />
                            )}
                        </div>
                    </div>
                )}

                {/* Package Versions Section */}
                <div className="border-t pt-6">
                    <h2 className="text-xl font-bold mb-4 flex items-center gap-2">
                        <Tag className="w-6 h-6" />
                        Package Versions ({packageVersions.length})
                    </h2>
                    {packageVersions.length === 0 ? (
                        <div className="text-center py-8 text-gray-500">
                            <Tag className="w-12 h-12 mx-auto mb-3 text-gray-400" />
                            <p>No versions found for this package</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {packageVersions.map((version) => {
                                const isActive = version.id === packageData.active_install_id;
                                return (
                                    <VersionCard
                                        key={version.id}
                                        version={version}
                                        isActive={isActive}
                                    />
                                );
                            })}
                        </div>
                    )}
                </div>

                {/* Artifacts Section */}
                <div className="border-t pt-6">
                    <h2 className="text-xl font-bold mb-4 flex items-center gap-2">
                        <Box className="w-6 h-6" />
                        Artifacts ({packageSpaces.length})
                    </h2>
                    {packageSpaces.length === 0 ? (
                        <div className="text-center py-8 text-gray-500">
                            <Box className="w-12 h-12 mx-auto mb-3 text-gray-400" />
                            <p>No artifacts found for this package</p>
                        </div>
                    ) : (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {packageSpaces.map((space, index) => {
                                const gradient = staticGradients[space.id % staticGradients.length];
                                return (
                                    <ArtifactCard
                                        key={space.id}
                                        space={space}
                                        gradient={gradient}
                                        isActive={space.id === spaceId}
                                    />
                                );
                            })}
                        </div>
                    )}
                </div>

                {/* Dev Token Section */}
                <div className="border-t pt-6">
                    <h2 className="text-xl font-bold mb-4 flex items-center gap-2">
                        <Key className="w-6 h-6" />
                        Development Token
                    </h2>
                    <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
                        <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
                            Generate a development token for CLI package push operations. 
                            This token allows you to update this package from the command line.
                        </p>
                        
                        {!devToken ? (
                            <button
                                onClick={handleGenerateDevToken}
                                disabled={generatingToken}
                                className="flex items-center gap-2 px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                <Key className="w-4 h-4" />
                                {generatingToken ? 'Generating...' : 'Generate Dev Token'}
                            </button>
                        ) : (
                            <div className="space-y-3">
                                <div className="flex items-center gap-2">
                                    <div className="flex-1 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg p-3 font-mono text-sm break-all">
                                        {devToken}
                                    </div>
                                    <button
                                        onClick={handleCopyToken}
                                        className="flex items-center gap-2 px-4 py-2 bg-gray-200 dark:bg-gray-600 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-300 dark:hover:bg-gray-500"
                                    >
                                        {copied ? (
                                            <>
                                                <Check className="w-4 h-4 text-green-600" />
                                                <span>Copied!</span>
                                            </>
                                        ) : (
                                            <>
                                                <Copy className="w-4 h-4" />
                                                <span>Copy</span>
                                            </>
                                        )}
                                    </button>
                                </div>
                                <button
                                    onClick={handleGenerateDevToken}
                                    disabled={generatingToken}
                                    className="text-sm text-blue-500 hover:text-blue-600 disabled:opacity-50"
                                >
                                    Generate New Token
                                </button>
                            </div>
                        )}
                    </div>
                </div>
            </div>
            </div>
        </WithAdminBodyLayout>
    );
};

interface DetailCardProps {
    icon: React.ReactNode;
    label: string;
    value: string;
    link?: string;
}

const DetailCard = ({ icon, label, value, link }: DetailCardProps) => {
    return (
        <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
            <div className="flex items-start gap-3">
                <div className="text-blue-500 dark:text-blue-400 mt-1">
                    {icon}
                </div>
                <div className="flex-1">
                    <p className="text-sm text-gray-500 dark:text-gray-400 mb-1">{label}</p>
                    {link ? (
                        <a
                            href={link}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-500 hover:text-blue-600 flex items-center gap-1"
                        >
                            {value}
                            <ExternalLink className="w-3 h-3" />
                        </a>
                    ) : (
                        <p className="text-gray-900 dark:text-white font-medium break-words">
                            {value}
                        </p>
                    )}
                </div>
            </div>
        </div>
    );
};

interface ArtifactCardProps {
    space: Space;
    gradient: string;
    isActive: boolean;
}

interface VersionCardProps {
    version: PackageVersion;
    isActive: boolean;
}

const VersionCard = ({ version, isActive }: VersionCardProps) => {
    return (
        <div className={`relative overflow-hidden rounded-lg bg-gray-50 dark:bg-gray-800 p-4 border-2 ${isActive ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20' : 'border-gray-200 dark:border-gray-700'}`}>
            <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between">
                    <span className="text-sm font-semibold text-gray-600 dark:text-gray-400">#{version.id}</span>
                    {isActive && (
                        <span className="bg-blue-500 text-white px-2 py-1 rounded text-xs font-semibold">
                            Active
                        </span>
                    )}
                </div>
                
                <div>
                    <h3 className="font-bold text-lg text-gray-900 dark:text-white">{version.name}</h3>
                    <div className="flex items-center gap-2 mt-2">
                        <span className="bg-blue-100 dark:bg-blue-900/40 text-blue-700 dark:text-blue-300 px-2 py-1 rounded text-xs">
                            v{version.version}
                        </span>
                        {version.format_version && (
                            <span className="bg-gray-200 dark:bg-gray-700 text-gray-700 dark:text-gray-300 px-2 py-1 rounded text-xs">
                                Format: {version.format_version}
                            </span>
                        )}
                    </div>
                </div>

                {version.info && (
                    <p className="text-sm text-gray-600 dark:text-gray-400 mt-2 line-clamp-2">
                        {version.info}
                    </p>
                )}

                <div className="flex items-center gap-4 text-xs mt-2">
                    {version.slug && (
                        <div className="flex items-center gap-1 text-gray-500 dark:text-gray-400">
                            <Code2 className="w-3 h-3" />
                            <span>{version.slug}</span>
                        </div>
                    )}
                    {version.license && (
                        <div className="flex items-center gap-1 text-gray-500 dark:text-gray-400">
                            <Scale className="w-3 h-3" />
                            <span>{version.license}</span>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
};

const ArtifactCard = ({ space, gradient, isActive }: ArtifactCardProps) => {
    return (
        <div className={`relative overflow-hidden rounded-lg bg-gradient-to-br ${gradient} p-4 text-white ${isActive ? 'ring-4 ring-yellow-400' : ''}`}>
            <div className="flex flex-col gap-2">
                <div className="flex items-center justify-between">
                    <span className="text-sm font-semibold">#{space.id}</span>
                    {isActive && (
                        <span className="bg-yellow-400 text-yellow-900 px-2 py-1 rounded text-xs font-semibold">
                            Current
                        </span>
                    )}
                </div>
                
                <div>
                    <h3 className="font-bold text-lg">{space.namespace_key}</h3>
                    <div className="flex items-center gap-2 mt-2">
                        <span className="bg-white/20 backdrop-blur-sm px-2 py-1 rounded text-xs">
                            {space.executor_type}
                        </span>
                        {space.sub_type && (
                            <span className="bg-white/20 backdrop-blur-sm px-2 py-1 rounded text-xs">
                                {space.sub_type}
                            </span>
                        )}
                    </div>
                </div>

                <div className="flex items-center gap-4 text-xs mt-2">
                    <div className="flex items-center gap-1">
                        <Box className="w-3 h-3" />
                        <span>{space.is_public ? 'Public' : 'Private'}</span>
                    </div>
                    <div className="flex items-center gap-1">
                        <Check className="w-3 h-3" />
                        <span>{space.is_initilized ? 'Initialized' : 'Not Initialized'}</span>
                    </div>
                </div>
            </div>
        </div>
    );
};
