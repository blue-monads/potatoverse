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
import { listInstalledSpaces, InstalledSpace, generatePackageDevToken, Package, Space } from '@/lib';
import { staticGradients } from '@/app/utils';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const packageId = searchParams.get('packageId');
    const spaceId = searchParams.get('id');
    const gapp = useGApp();

    if (!packageId) {
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

    return <PackageAbout packageId={parseInt(packageId)} spaceId={spaceId ? parseInt(spaceId) : undefined} />;
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

    const loader = useSimpleDataLoader<InstalledSpace>({
        loader: listInstalledSpaces,
        ready: gapp.isInitialized,
    });

    const packageData = loader.data?.packages.find(pkg => pkg.id === packageId);
    const packageSpaces = loader.data?.spaces.filter(space => space.package_id === packageId) || [];

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
                            <p className="text-blue-100 mb-4">{packageData.info || 'No description available'}</p>
                            <div className="flex items-center gap-4">
                                <span className="bg-white/20 px-3 py-1 rounded-full text-sm">
                                    v{packageData.version}
                                </span>
                                <span className="bg-white/20 px-3 py-1 rounded-full text-sm">
                                    ID: {packageData.id}
                                </span>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Package Details Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {/* Left Column */}
                    <div className="space-y-4">
                        <DetailCard
                            icon={<Code2 className="w-5 h-5" />}
                            label="Slug"
                            value={packageData.slug || 'N/A'}
                        />
                        <DetailCard
                            icon={<Tag className="w-5 h-5" />}
                            label="Tags"
                            value={packageData.tags || 'No tags'}
                        />
                        <DetailCard
                            icon={<Scale className="w-5 h-5" />}
                            label="License"
                            value={packageData.type || 'N/A'}
                        />
                    </div>

                    {/* Right Column */}
                    <div className="space-y-4">
                        <DetailCard
                            icon={<Calendar className="w-5 h-5" />}
                            label="Format Version"
                            value={packageData.version || 'N/A'}
                        />
                        <DetailCard
                            icon={<Info className="w-5 h-5" />}
                            label="Description"
                            value={packageData.description || packageData.info || 'No description'}
                        />
                    </div>
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
