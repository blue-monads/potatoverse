"use client";
import React, { useState } from 'react';
import { Blocks, Trash2, Search, Package, Layers, Plug, FileCode, CheckCircle2, AlertCircle, X, ExternalLink, Settings } from 'lucide-react';
import { useSearchParams } from 'next/navigation';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import {
    listSpacePlugins,
    listAvailablePlugins,
    createSpacePlugin,
    deleteSpacePlugin,
    updateSpacePlugin,
    SpacePlugin,
    AvailablePlugin,
} from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';

export default function Page() {
    const searchParams = useSearchParams();
    const installId = searchParams.get('install_id');
    const spaceId = searchParams.get('space_id');

    if (!installId) {
        return (
            <div className="p-8 text-center text-gray-500">
                Install ID not provided
            </div>
        );
    }

    return (
        <PluginsListingPage
            installId={parseInt(installId, 10)}
            spaceId={spaceId ? parseInt(spaceId, 10) : undefined}
        />
    );
}

interface PluginsListingPageProps {
    installId: number;
    spaceId?: number;
}

const PluginsListingPage = ({ installId, spaceId }: PluginsListingPageProps) => {
    const [searchTerm, setSearchTerm] = useState('');
    const [isWizardOpen, setIsWizardOpen] = useState(false);
    const [editingPlugin, setEditingPlugin] = useState<SpacePlugin | null>(null);
    const [editMetaValue, setEditMetaValue] = useState('');

    const pluginsLoader = useSimpleDataLoader<SpacePlugin[]>({
        loader: () => listSpacePlugins(installId, spaceId),
        ready: !!installId,
        dependencies: [installId, spaceId],
    });

    const availLoader = useSimpleDataLoader<AvailablePlugin[]>({
        loader: () => listAvailablePlugins(installId, spaceId),
        ready: isWizardOpen && !!installId,
        dependencies: [installId, spaceId, isWizardOpen],
    });

    const filteredPlugins = pluginsLoader.data?.filter(p => {
        if (!searchTerm) return true;
        const term = searchTerm.toLowerCase();
        return (
            (p.target_namespace_key && p.target_namespace_key.toLowerCase().includes(term)) ||
            (p.target_package_name && p.target_package_name.toLowerCase().includes(term)) ||
            (p.target_loader_script && p.target_loader_script.toLowerCase().includes(term)) ||
            p.target_space_id.toString().includes(term) ||
            p.target_install_id.toString().includes(term)
        );
    }) || [];

    const handleDelete = async (pluginId: number) => {
        try {
            await deleteSpacePlugin(installId, pluginId);
            pluginsLoader.reload();
            if (isWizardOpen) {
                availLoader.reload();
            }
        } catch (error) {
            console.error('Failed to unplug plugin:', error);
            alert('Failed to unplug plugin: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    const handleSaveEdit = async () => {
        if (!editingPlugin) return;
        try {
            await updateSpacePlugin(installId, editingPlugin.id, { extrameta: editMetaValue });
            setEditingPlugin(null);
            pluginsLoader.reload();
        } catch (error) {
            console.error('Failed to update plugin:', error);
            alert('Failed to update plugin: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Blocks}
            name="Plugins"
            description="Manage connected AppPlugins for this space"
            rightContent={
                <AddButton
                    name="+ Plugin"
                    onClick={() => setIsWizardOpen(true)}
                />
            }
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
                <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                        <div className="w-3 h-3 bg-purple-500 rounded-full"></div>
                        <h2 className="text-xl font-bold">Connected Plugins</h2>
                        <span className="text-xs bg-purple-100 text-purple-800 font-semibold px-2 py-0.5 rounded-full">
                            {filteredPlugins.length}
                        </span>
                    </div>
                </div>

                <div className="bg-white rounded-lg shadow overflow-hidden">
                    <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-gray-200">
                            <thead className="bg-gray-50">
                                <tr>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Plugin Namespace
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Package className="w-4 h-4" />
                                            Target Package
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <Layers className="w-4 h-4" />
                                            Target Space ID
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        <div className="flex items-center gap-2">
                                            <FileCode className="w-4 h-4" />
                                            Loader Script
                                        </div>
                                    </th>
                                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Configuration
                                    </th>
                                    <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                        Actions
                                    </th>
                                </tr>
                            </thead>
                            <tbody className="bg-white divide-y divide-gray-200">
                                {pluginsLoader.loading ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-8 text-center text-gray-500">
                                            Loading plugins...
                                        </td>
                                    </tr>
                                ) : filteredPlugins.length === 0 ? (
                                    <tr>
                                        <td colSpan={6} className="px-6 py-12 text-center">
                                            <div className="flex flex-col items-center justify-center space-y-3">
                                                <Blocks className="w-12 h-12 text-gray-300" />
                                                <p className="text-gray-500 font-medium">No plugins connected to this space</p>
                                                <p className="text-sm text-gray-400">
                                                    Click &quot;+ Plugin&quot; to browse and plug in available AppPlugins.
                                                </p>
                                                <button
                                                    onClick={() => setIsWizardOpen(true)}
                                                    className="btn btn-sm preset-filled text-white bg-blue-600 hover:bg-blue-700 mt-2"
                                                >
                                                    Add Plugin
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ) : (
                                    filteredPlugins.map(plugin => (
                                        <tr key={plugin.id} className="hover:bg-gray-50">
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="flex items-center gap-3">
                                                    <div className="p-2 bg-purple-50 text-purple-600 rounded-lg">
                                                        <Plug className="w-5 h-5" />
                                                    </div>
                                                    <div>
                                                        <div className="text-sm font-semibold text-gray-900">
                                                            {plugin.target_namespace_key || `Space #${plugin.target_space_id}`}
                                                        </div>
                                                        <div className="text-xs text-gray-400">
                                                            ID: {plugin.id}
                                                        </div>
                                                    </div>
                                                </div>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                <div className="text-sm text-gray-900 font-medium">
                                                    {plugin.target_package_name || `Install #${plugin.target_install_id}`}
                                                </div>
                                                {plugin.target_package_version && (
                                                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-100 text-gray-800">
                                                        v{plugin.target_package_version}
                                                    </span>
                                                )}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                                {plugin.target_space_id}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap">
                                                {plugin.target_loader_script ? (
                                                    <code className="text-xs bg-gray-100 text-gray-800 px-2 py-1 rounded font-mono">
                                                        {plugin.target_loader_script}
                                                    </code>
                                                ) : (
                                                    <span className="text-xs text-gray-400 italic">None</span>
                                                )}
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                                <button
                                                    onClick={() => {
                                                        setEditingPlugin(plugin);
                                                        setEditMetaValue(plugin.extrameta || '{}');
                                                    }}
                                                    className="inline-flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 cursor-pointer"
                                                >
                                                    <Settings className="w-3.5 h-3.5" />
                                                    {plugin.extrameta && plugin.extrameta !== '{}' ? 'Configured' : 'Configure'}
                                                </button>
                                            </td>
                                            <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                                <button
                                                    onClick={() => {
                                                        if (confirm(`Are you sure you want to disconnect ${plugin.target_namespace_key || 'this plugin'}?`)) {
                                                            handleDelete(plugin.id);
                                                        }
                                                    }}
                                                    className="text-red-600 hover:text-red-900 inline-flex items-center gap-1 cursor-pointer"
                                                    title="Disconnect Plugin"
                                                >
                                                    <Trash2 className="w-4 h-4" />
                                                    <span>Unplug</span>
                                                </button>
                                            </td>
                                        </tr>
                                    ))
                                )}
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            {/* Add Plugin Wizard Modal */}
            {isWizardOpen && (
                <AddPluginWizardModal
                    installId={installId}
                    spaceId={spaceId}
                    availablePlugins={availLoader.data || []}
                    loading={availLoader.loading}
                    onClose={() => setIsWizardOpen(false)}
                    onPluginAdded={() => {
                        pluginsLoader.reload();
                        availLoader.reload();
                    }}
                />
            )}

            {/* Edit Meta Modal */}
            {editingPlugin && (
                <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
                    <div className="bg-white rounded-xl shadow-2xl max-w-lg w-full p-6 space-y-4">
                        <div className="flex items-center justify-between border-b pb-3">
                            <h3 className="text-lg font-bold text-gray-900">
                                Configure Plugin: {editingPlugin.target_namespace_key}
                            </h3>
                            <button
                                onClick={() => setEditingPlugin(null)}
                                className="text-gray-400 hover:text-gray-600"
                            >
                                <X className="w-5 h-5" />
                            </button>
                        </div>
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-1">
                                Extra Metadata (JSON)
                            </label>
                            <textarea
                                value={editMetaValue}
                                onChange={e => setEditMetaValue(e.target.value)}
                                rows={5}
                                className="w-full font-mono text-sm border border-gray-300 rounded-lg p-2.5 focus:ring-blue-500 focus:border-blue-500"
                            />
                        </div>
                        <div className="flex justify-end gap-2 pt-2">
                            <button
                                onClick={() => setEditingPlugin(null)}
                                className="btn btn-sm bg-gray-100 hover:bg-gray-200 text-gray-700"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={handleSaveEdit}
                                className="btn btn-sm bg-blue-600 hover:bg-blue-700 text-white font-medium"
                            >
                                Save
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </WithAdminBodyLayout>
    );
};

interface AddPluginWizardModalProps {
    installId: number;
    spaceId?: number;
    availablePlugins: AvailablePlugin[];
    loading: boolean;
    onClose: () => void;
    onPluginAdded: () => void;
}

const AddPluginWizardModal = ({
    installId,
    spaceId,
    availablePlugins,
    loading,
    onClose,
    onPluginAdded,
}: AddPluginWizardModalProps) => {
    const [filterQuery, setFilterQuery] = useState('');
    const [connectingId, setConnectingId] = useState<number | null>(null);

    const filtered = availablePlugins.filter(p => {
        if (!filterQuery) return true;
        const q = filterQuery.toLowerCase();
        return (
            p.namespace_key.toLowerCase().includes(q) ||
            p.package_name.toLowerCase().includes(q) ||
            (p.package_info && p.package_info.toLowerCase().includes(q)) ||
            (p.package_author && p.package_author.toLowerCase().includes(q))
        );
    });

    const handleConnect = async (plugin: AvailablePlugin) => {
        setConnectingId(plugin.space_id);
        try {
            await createSpacePlugin(installId, spaceId, {
                target_install_id: plugin.install_id,
                target_space_id: plugin.space_id,
                source_space_id: spaceId,
                extrameta: '{}',
            });
            onPluginAdded();
        } catch (error) {
            console.error('Failed to plug plugin:', error);
            alert('Failed to plug plugin: ' + ((error as any)?.response?.data?.error || (error as any)?.message));
        } finally {
            setConnectingId(null);
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
            <div className="bg-white rounded-xl shadow-2xl max-w-3xl w-full max-h-[85vh] flex flex-col overflow-hidden">
                {/* Modal Header */}
                <div className="p-6 border-b border-gray-200 flex items-center justify-between">
                    <div>
                        <h3 className="text-xl font-bold text-gray-900 flex items-center gap-2">
                            <Plug className="w-5 h-5 text-purple-600" />
                            Add Plugin Wizard
                        </h3>
                        <p className="text-sm text-gray-500 mt-1">
                            Select an available AppPlugin to attach to this Space App
                        </p>
                    </div>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-600 p-1 rounded-lg"
                    >
                        <X className="w-6 h-6" />
                    </button>
                </div>

                {/* Search Bar */}
                <div className="px-6 py-3 border-b border-gray-100 bg-gray-50">
                    <div className="relative">
                        <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                        <input
                            type="text"
                            placeholder="Search by plugin namespace, package, author..."
                            value={filterQuery}
                            onChange={e => setFilterQuery(e.target.value)}
                            className="w-full pl-9 pr-4 py-2 text-sm bg-white border border-gray-300 rounded-lg focus:ring-blue-500 focus:border-blue-500"
                        />
                    </div>
                </div>

                {/* Plugin List */}
                <div className="p-6 overflow-y-auto space-y-4 flex-1">
                    {loading ? (
                        <div className="py-12 text-center text-gray-500">
                            Loading available AppPlugins...
                        </div>
                    ) : filtered.length === 0 ? (
                        <div className="py-12 text-center">
                            <Blocks className="w-12 h-12 text-gray-300 mx-auto mb-3" />
                            <p className="text-gray-600 font-medium">No AppPlugins available</p>
                            <p className="text-sm text-gray-400 mt-1">
                                {filterQuery
                                    ? 'No plugins matched your search.'
                                    : 'There are no other packages installed that define spaces with type "AppPlugin".'}
                            </p>
                        </div>
                    ) : (
                        filtered.map(plugin => (
                            <div
                                key={plugin.space_id}
                                className={`p-4 border rounded-xl flex items-center justify-between gap-4 transition-all ${
                                    plugin.is_already_plugged
                                        ? 'bg-gray-50 border-gray-200 opacity-75'
                                        : 'bg-white border-gray-200 hover:border-purple-300 hover:shadow-sm'
                                }`}
                            >
                                <div className="space-y-1 min-w-0 flex-1">
                                    <div className="flex items-center gap-2">
                                        <span className="font-semibold text-gray-900 text-sm truncate">
                                            {plugin.namespace_key}
                                        </span>
                                        {plugin.is_already_plugged && (
                                            <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-800 font-medium">
                                                <CheckCircle2 className="w-3 h-3" /> Plugged
                                            </span>
                                        )}
                                        {plugin.loader_script && (
                                            <span className="inline-flex items-center gap-1 text-xs px-2 py-0.5 rounded bg-gray-100 text-gray-600 font-mono">
                                                <FileCode className="w-3 h-3" /> {plugin.loader_script}
                                            </span>
                                        )}
                                    </div>
                                    <div className="text-xs text-gray-500 flex items-center gap-3">
                                        <span className="font-medium text-gray-700">{plugin.package_name}</span>
                                        {plugin.package_version && <span>v{plugin.package_version}</span>}
                                        {plugin.package_author && <span>by {plugin.package_author}</span>}
                                    </div>
                                    {plugin.package_info && (
                                        <p className="text-xs text-gray-500 line-clamp-1">{plugin.package_info}</p>
                                    )}
                                </div>

                                <div>
                                    {plugin.is_already_plugged ? (
                                        <button
                                            disabled
                                            className="btn btn-sm bg-gray-100 text-gray-400 cursor-not-allowed font-medium px-4"
                                        >
                                            Connected
                                        </button>
                                    ) : (
                                        <button
                                            onClick={() => handleConnect(plugin)}
                                            disabled={connectingId === plugin.space_id}
                                            className="btn btn-sm preset-filled text-white bg-purple-600 hover:bg-purple-700 font-medium px-4 flex items-center gap-1 cursor-pointer"
                                        >
                                            <Plug className="w-3.5 h-3.5" />
                                            {connectingId === plugin.space_id ? 'Connecting...' : 'Connect'}
                                        </button>
                                    )}
                                </div>
                            </div>
                        ))
                    )}
                </div>

                {/* Footer */}
                <div className="p-4 border-t border-gray-200 bg-gray-50 flex justify-end">
                    <button
                        onClick={onClose}
                        className="btn btn-sm bg-gray-200 hover:bg-gray-300 text-gray-700 font-medium"
                    >
                        Done
                    </button>
                </div>
            </div>
        </div>
    );
};
