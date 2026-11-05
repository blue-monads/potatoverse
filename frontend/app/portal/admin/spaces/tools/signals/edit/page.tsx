"use client";
import React from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import SignalEditor from '../components/SignalEditor';
import { getSignal, updateSignal, Signal } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');
    const signalId = searchParams.get('signal_id');

    if (!installId || !signalId) {
        return <div>Install ID or Signal ID not provided</div>;
    }

    const loader = useSimpleDataLoader<Signal>({
        loader: () => getSignal(parseInt(installId), parseInt(signalId)),
        ready: true,
        dependencies: [installId, signalId],
    });

    const handleSave = async (data: any) => {
        try {
            await updateSignal(parseInt(installId), parseInt(signalId), data);
            const params = new URLSearchParams(searchParams.toString());
            params.delete('signal_id');
            router.push(`/portal/admin/spaces/tools/signals?${params.toString()}`);
        } catch (error) {
            console.error('Failed to update signal:', error);
            alert('Failed to update signal: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
            throw error;
        }
    };

    const handleBack = () => {
        const params = new URLSearchParams(searchParams.toString());
        params.delete('signal_id');
        router.push(`/portal/admin/spaces/tools/signals?${params.toString()}`);
    };

    if (loader.loading) {
        return (
            <div className="flex items-center justify-center h-64">
                <div className="text-center">
                    <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mx-auto mb-4"></div>
                    <p className="text-gray-500">Loading signal...</p>
                </div>
            </div>
        );
    }

    if (!loader.data) {
        return <div>Signal not found</div>;
    }

    return (
        <SignalEditor
            onSave={handleSave}
            onBack={handleBack}
            initialData={loader.data}
            installId={parseInt(installId)}
            spaceId={spaceId ? parseInt(spaceId) : undefined}
        />
    );
}
