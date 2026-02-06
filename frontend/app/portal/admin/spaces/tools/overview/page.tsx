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
    Info,
    FileIcon
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
   
    if (!installId) {
        return (
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

    const loader = useSimpleDataLoader<InstalledPackageInfo>({
        loader: () => getInstalledPackageInfo(packageId),
        ready: gapp.isInitialized,
    });

    const packageData = loader.data?.installed_package;
    const packageVersions = loader.data?.package_versions || [];
    const packageSpaces = loader.data?.spaces || [];
    const activeVersion = packageData ? packageVersions.find(v => v.id === packageData.active_install_id) : null;


    if (loader.loading) {
        return (
            <p className="text-gray-500">Loading package information...</p>
        );
    }

    if (!packageData) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="text-center">
                    <PackageIcon className="w-16 h-16 text-gray-400 mx-auto mb-4" />
                    <p className="text-gray-500">Package not found</p>
                </div>
            </div>
        );
    }

    return (
        <WithAdminBodyLayout
            Icon={PackageIcon}
            name="Package Overview"
            description={`Overview for ${packageData.name}`}
            variant="none"
        >
            <div className="max-w-7xl mx-auto w-full px-6 py-8">
                <div className="card p-6 flex flex-col gap-6">
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
        <div className="bg-gray-50 rounded-lg p-4">
            <div className="flex items-start gap-3">
                <div className="text-blue-500 mt-1">
                    {icon}
                </div>
                <div className="flex-1">
                    <p className="text-sm text-gray-500 mb-1">{label}</p>
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
                        <p className="text-gray-900 font-medium break-words">
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
