"use client";
import React, { useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { ArrowLeft } from 'lucide-react';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import { 
    createSpaceCapability,
    listCapabilityTypes,
    CapabilityDefinition
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import CapNewEditor from '../sub/CapNewEditor';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');
    const capabilityType = searchParams.get('capability_type');

    if (!installId) {
        return <div>Install ID not provided</div>;
    }

    const capabilityTypesLoader = useSimpleDataLoader<CapabilityDefinition[]>({
        loader: () => listCapabilityTypes(),
        ready: true,
    });

    const handleSave = async (data: {
        name: string;
        capability_type: string;
        space_id?: number;
        options?: any;
        extrameta?: any;
    }) => {
        try {
            await createSpaceCapability(parseInt(installId), data);
            const params = new URLSearchParams();
            params.set('install_id', installId);
            if (spaceId) params.set('space_id', spaceId);
            router.push(`/portal/admin/spaces/tools/capabilities?${params.toString()}`);
        } catch (error) {
            console.error('Failed to create capability:', error);
            alert('Failed to create capability: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
            throw error;
        }
    };

    const handleBack = () => {
        const params = new URLSearchParams();
        params.set('install_id', installId);
        if (spaceId) params.set('space_id', spaceId);
        router.push(`/portal/admin/spaces/tools/capabilities?${params.toString()}`);
    };

    return (
        <WithAdminBodyLayout
            Icon={ArrowLeft}
            name="Create Capability"
            description="Add a new capability to this package or space"
            rightContent={
                <button
                    onClick={handleBack}
                    className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-900"
                >
                    <ArrowLeft className="w-4 h-4" />
                    Back
                </button>
            }
        >
            <div className="max-w-4xl mx-auto px-6 py-8 w-full">
                {capabilityTypesLoader.loading ? (
                    <div className="flex items-center justify-center h-64">
                        <div className="text-center">
                            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                            <p className="text-gray-500">Loading capability types...</p>
                        </div>
                    </div>
                ) : (
                    <CapNewEditor
                        capabilityTypes={capabilityTypesLoader.data || []}
                        defaultSpaceId={spaceId ? parseInt(spaceId) : 0}
                        onSave={handleSave}
                        onCancel={handleBack}
                        capabilityType={capabilityType || ''}
                    />
                )}
            </div>
        </WithAdminBodyLayout>
    );
}

