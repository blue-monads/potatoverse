"use client";
import React, { useEffect, useMemo, useRef, useState } from 'react';
import { Search, Filter, ArrowUpDown, Heart, Users, Zap, Image, Box, Octagon, SquareUserRound, BadgeDollarSign, BookOpenText, BookHeart, BriefcaseBusiness, Drama, Bolt, CloudLightning, ScrollText, Files, Grid2x2Plus, Cog, Trash2Icon, FileCode2, BoltIcon, List, Plug, Globe, Wrench, AlertTriangle, X } from 'lucide-react';
import { createPortal } from 'react-dom';
import WithAdminBodyLayout from '@/contain/Layouts/WithAdminBodyLayout';
import BigSearchBar from '@/contain/compo/BigSearchBar';
import { AddButton } from '@/contain/AddButton';
import { GAppStateHandle, ModalHandle, useGApp } from '@/hooks';
import { Tabs } from '@skeletonlabs/skeleton-react';
import { deletePackage, formatSpace, FormattedSpace, InstalledSpace, installPackage, installPackageZip, isWildcardHost, listInstalledSpaces, Package, Space } from '@/lib';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { staticGradients } from '@/app/utils';
import { useRouter } from 'next/navigation';
import useFavorites from '@/hooks/useFavorites/useFavorites';
import { deriveHostAndIframeSrc } from '../exec/hostSrc';
import SimpleLoader from '@/contain/SimpleLoader/SimpleLoader';



export default function Page() {
    return (<>
        <SpacesDirectory />
    </>)
}



const SpacesDirectory = () => {
    const [searchTerm, setSearchTerm] = useState('');
    const [selectedFilter, setSelectedFilter] = useState('Relevance');
    const [showPlugins, setShowPlugins] = useState(false);
    const [showRootApps, setShowRootApps] = useState(false);
    const [viewMode, setViewMode] = useState<'cards' | 'list'>('cards');
    const gapp = useGApp();
    const router = useRouter();
    const { favorites, addFavorite, removeFavorite } = useFavorites();
    const [formattedSpaces, setFormattedSpaces] = useState<FormattedSpace[]>([]);
    const [showDomainInfo, setShowDomainInfo] = useState(false);

    useEffect(() => {
        try {
            const savedMode = localStorage.getItem('spaces_view_mode');
            if (savedMode === 'cards' || savedMode === 'list') {
                setViewMode(savedMode);
            }
            const savedShowPlugins = localStorage.getItem('spaces_show_plugins');
            if (savedShowPlugins !== null) {
                setShowPlugins(savedShowPlugins === 'true');
            }
            const savedShowRootApps = localStorage.getItem('spaces_show_root_apps');
            if (savedShowRootApps !== null) {
                setShowRootApps(savedShowRootApps === 'true');
            }
        } catch (e) {
            // ignore localStorage access errors
        }
    }, []);

    const changeViewMode = (mode: 'cards' | 'list') => {
        setViewMode(mode);
        try {
            localStorage.setItem('spaces_view_mode', mode);
        } catch (e) {
            // ignore localStorage access errors
        }
    };

    const changeShowPlugins = (val: boolean) => {
        setShowPlugins(val);
        try {
            localStorage.setItem('spaces_show_plugins', String(val));
        } catch (e) {
            // ignore localStorage access errors
        }
    };

    const changeShowRootApps = (val: boolean) => {
        setShowRootApps(val);
        try {
            localStorage.setItem('spaces_show_root_apps', String(val));
        } catch (e) {
            // ignore localStorage access errors
        }
    };

    const loader = useSimpleDataLoader<InstalledSpace>({
        loader: listInstalledSpaces,
        ready: gapp.isInitialized,
    });

    useEffect(() => {
        if (loader.data) {
            const nextFormattedSpaces = formatSpace(loader.data);
            setFormattedSpaces(nextFormattedSpaces);
        }
    }, [loader.data]);

    const sortOptions = [
        'Relevance',
        'Recently Created',
        'Recently Updated',
        'Installed Date',
        'By Usage'
    ];

    const [isDropdownOpen, setIsDropdownOpen] = useState(false);

    const displaySpaces = useMemo(() => {
        let spaces = formattedSpaces;

        if (!showPlugins) {
            spaces = spaces.filter((s) => !s.is_plugin);
        }

        if (!showRootApps) {
            spaces = spaces.filter((s) => s.space_type !== 'RootApp');
        }

        if (searchTerm.trim()) {
            const term = searchTerm.toLowerCase();
            spaces = spaces.filter((s) => {
                const matchesName = s.package_name.toLowerCase().includes(term);
                const matchesNs = s.namespace_key.toLowerCase().includes(term);
                const matchesInfo = s.package_info?.toLowerCase().includes(term);
                const matchesAuthor = s.package_author?.toLowerCase().includes(term);
                const matchesId = s.space_id.toString().includes(term);
                return matchesName || matchesNs || matchesInfo || matchesAuthor || matchesId;
            });
        }

        if (selectedFilter === 'Recently Created' || selectedFilter === 'Installed Date' || selectedFilter === 'Recently Updated') {
            spaces = [...spaces].sort((a, b) => b.space_id - a.space_id);
        }

        return spaces;
    }, [formattedSpaces, showPlugins, showRootApps, searchTerm, selectedFilter]);

    const packageGroups = useMemo(() => {
        if (!loader.data) return [];

        const spacesByInstallId = new Map<number, FormattedSpace[]>();
        for (const space of displaySpaces) {
            const list = spacesByInstallId.get(space.install_id) || [];
            list.push(space);
            spacesByInstallId.set(space.install_id, list);
        }

        const groups: { pkg?: Package; installId: number; spaces: FormattedSpace[] }[] = [];
        const seenInstallIds = new Set<number>();

        for (const pkg of loader.data.packages) {
            const spaces = spacesByInstallId.get(pkg.install_id);
            if (spaces && spaces.length > 0) {
                groups.push({
                    pkg,
                    installId: pkg.install_id,
                    spaces,
                });
                seenInstallIds.add(pkg.install_id);
            }
        }

        for (const [installId, spaces] of spacesByInstallId.entries()) {
            if (!seenInstallIds.has(installId)) {
                groups.push({
                    pkg: undefined,
                    installId,
                    spaces,
                });
            }
        }

        return groups;
    }, [loader.data, displaySpaces]);

    const configuredHosts = useMemo(() => {
        if (loader.data?.hosts && loader.data.hosts.length > 0) {
            return loader.data.hosts;
        }
        if (typeof window !== 'undefined' && (window as any).__potato_attrs__?.site_hosts) {
            return ((window as any).__potato_attrs__.site_hosts as string)
                .split(',')
                .map((h) => h.trim())
                .filter(Boolean);
        }
        return [];
    }, [loader.data?.hosts]);

    const isCurrentHostWildcard = useMemo(() => {
        if (typeof window === 'undefined') return false;
        return isWildcardHost(configuredHosts, window.location.host);
    }, [configuredHosts]);

    const spaceConflictMap = useMemo(() => {
        const result = new Map<number, { isConflict: boolean; winnerSpaceId: number }>();
        if (isCurrentHostWildcard) {
            return result;
        }

        const spacesByNs = new Map<string, number[]>();
        const allSpaces = loader.data?.spaces || [];

        for (const s of allSpaces) {
            const list = spacesByNs.get(s.namespace_key) || [];
            list.push(s.id);
            spacesByNs.set(s.namespace_key, list);
        }

        for (const [_, ids] of spacesByNs.entries()) {
            if (ids.length <= 1) continue;
            // On non-wildcard domains, lowest space ID wins the namespace
            const minId = Math.min(...ids);
            for (const id of ids) {
                if (id !== minId) {
                    result.set(id, { isConflict: true, winnerSpaceId: minId });
                }
            }
        }

        return result;
    }, [isCurrentHostWildcard, loader.data?.spaces]);

    const handleAction = async (action: string, space: FormattedSpace) => {
        const installId = space.install_id;
        const spaceId = space.space_id;
        const namespaceKey = space.namespace_key;
        const packageVersionId = space.package_version_id;

        if (action === "run") {
            const conflict = spaceConflictMap.get(spaceId);
            if (conflict?.isConflict) {
                return;
            }
            if (!spaceId) {
                return;
            }
            const hostsrc = await deriveHostAndIframeSrc(namespaceKey, String(spaceId));
            if (!hostsrc) {
                return;
            }
            window.open(hostsrc, '_blank');
            return;
        }

        const params = new URLSearchParams();
        params.set("install_id", installId.toString());
        params.set("space_id", spaceId.toString());
        params.set("namespace_key", namespaceKey);
        params.set("nskey", namespaceKey);
        params.set("package_version_id", packageVersionId.toString());

        if (action === "delete") {
            gapp.modal.openModal({
                title: "Delete Space",
                content: (
                    <div className="space-y-4">
                        <div className="flex items-center gap-3">
                            <div className="w-10 h-10 bg-red-100 rounded-full flex items-center justify-center">
                                <Trash2Icon className="w-5 h-5 text-red-600" />
                            </div>
                            <div>
                                <h3 className="text-lg font-semibold text-gray-900">
                                    Are you sure you want to delete this space?
                                </h3>
                                <p className="text-sm text-gray-600">
                                    This action cannot be undone. All data associated with this space will be permanently removed.
                                </p>
                            </div>
                        </div>

                        <div className="bg-gray-50 p-4 rounded-lg">
                            <div className="flex items-center gap-3">
                                <div className={`w-8 h-8 rounded-lg bg-gradient-to-br ${space.gradient} flex items-center justify-center text-white text-sm font-semibold`}>
                                    #{space.space_id}
                                </div>
                                <div>
                                    <p className="font-medium text-gray-900">{space.package_name}</p>
                                    <p className="text-sm text-gray-600">{space.package_info}</p>
                                </div>
                            </div>
                        </div>

                        <div className="flex gap-3 justify-end">
                            <button
                                onClick={() => gapp.modal.closeModal()}
                                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors cursor-pointer"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={async () => {
                                    try {
                                        await deletePackage(installId);
                                        loader.reload();
                                        gapp.modal.closeModal();
                                    } catch (error) {
                                        console.error('Failed to delete space:', error);
                                    }
                                }}
                                className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg transition-colors cursor-pointer"
                            >
                                Delete Space
                            </button>
                        </div>
                    </div>
                ),
                size: "md"
            });
        } else if (action === "tools") {
            router.push(`/portal/admin/spaces/tools/overview?${params.toString()}`);
        }
    };

    return (
        <WithAdminBodyLayout
            Icon={Box}
            name="Spaces"
            description="Your App Directory"
            rightContent={
                <AddButton
                    name="+ Space"
                    onClick={() => {
                        router.push('/portal/admin/store');
                    }}
                />
            }
        >
            <BigSearchBar
                searchText={searchTerm}
                setSearchText={setSearchTerm}
            />

            <div className="max-w-7xl mx-auto px-6 py-8 w-full">
                <div className="mb-8">
                    <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 mb-6">
                        <div className="flex items-center gap-3">
                            <div className="w-3 h-3 bg-orange-500 rounded-full"></div>
                            <h2 className="text-xl font-bold">Installed Spaces</h2>
                            <span className="text-xs px-2.5 py-0.5 bg-gray-100 text-gray-600 rounded-full font-medium">
                                {displaySpaces.length}
                            </span>
                            {isCurrentHostWildcard ? (
                                <span
                                    className="text-xs px-2.5 py-0.5 bg-emerald-50 text-emerald-700 border border-emerald-200 rounded-full font-medium"
                                    title="Current domain has wildcard support (*.domain). Spaces run in isolated subdomains."
                                >
                                    Wildcard Domain
                                </span>
                            ) : (
                                <button
                                    type="button"
                                    onClick={() => setShowDomainInfo((prev) => !prev)}
                                    className={`text-xs px-2.5 py-0.5 rounded-full font-medium border transition-colors cursor-pointer flex items-center gap-1.5 ${
                                        showDomainInfo
                                            ? 'bg-amber-200/90 text-amber-900 border-amber-400 ring-2 ring-amber-300'
                                            : 'bg-amber-50 text-amber-700 border-amber-200 hover:bg-amber-100'
                                    }`}
                                    title="Click to view details about non-wildcard domain behavior and namespace collisions"
                                >
                                    {spaceConflictMap.size > 0 && (
                                        <AlertTriangle className="w-3 h-3 text-amber-600" />
                                    )}
                                    <span>Non-wildcard Domain</span>
                                </button>
                            )}
                        </div>

                        <div className="flex items-center gap-3 flex-wrap">
                            {/* Toggle Show Plugins */}
                            <label className={`flex items-center gap-2 px-3 py-2 border rounded-lg text-sm transition-colors cursor-pointer select-none ${
                                showPlugins
                                    ? 'bg-purple-50 border-purple-300 text-purple-700'
                                    : 'border-gray-300 hover:bg-gray-50 text-gray-700'
                            }`}>
                                <input
                                    type="checkbox"
                                    checked={showPlugins}
                                    onChange={(e) => changeShowPlugins(e.target.checked)}
                                    className="w-4 h-4 text-purple-600 rounded border-gray-300 focus:ring-purple-500 cursor-pointer"
                                />
                                <span className="font-medium flex items-center gap-1.5">
                                    <Plug className="w-4 h-4" />
                                    Show plugins
                                </span>
                            </label>

                            {/* Toggle Show Root Apps */}
                            <label className={`flex items-center gap-2 px-3 py-2 border rounded-lg text-sm transition-colors cursor-pointer select-none ${
                                showRootApps
                                    ? 'bg-indigo-50 border-indigo-300 text-indigo-700'
                                    : 'border-gray-300 hover:bg-gray-50 text-gray-700'
                            }`}>
                                <input
                                    type="checkbox"
                                    checked={showRootApps}
                                    onChange={(e) => changeShowRootApps(e.target.checked)}
                                    className="w-4 h-4 text-indigo-600 rounded border-gray-300 focus:ring-indigo-500 cursor-pointer"
                                />
                                <span className="font-medium flex items-center gap-1.5">
                                    <Globe className="w-4 h-4" />
                                    Show root apps
                                </span>
                            </label>

                            {/* View Switcher: Card vs List */}
                            <div className="flex items-center border border-gray-300 rounded-lg p-0.5 bg-gray-50">
                                <button
                                    type="button"
                                    onClick={() => changeViewMode('cards')}
                                    className={`px-3 py-1.5 rounded-md flex items-center gap-1.5 text-xs font-medium transition-colors cursor-pointer ${
                                        viewMode === 'cards'
                                            ? 'bg-white shadow-xs text-blue-600 font-semibold'
                                            : 'text-gray-600 hover:text-gray-900'
                                    }`}
                                    title="Card View"
                                >
                                    <Grid2x2Plus className="w-4 h-4" />
                                    <span>Cards</span>
                                </button>
                                <button
                                    type="button"
                                    onClick={() => changeViewMode('list')}
                                    className={`px-3 py-1.5 rounded-md flex items-center gap-1.5 text-xs font-medium transition-colors cursor-pointer ${
                                        viewMode === 'list'
                                            ? 'bg-white shadow-xs text-blue-600 font-semibold'
                                            : 'text-gray-600 hover:text-gray-900'
                                    }`}
                                    title="List View"
                                >
                                    <List className="w-4 h-4" />
                                    <span>List</span>
                                </button>
                            </div>

                            {/* Sort dropdown */}
                            <div className="relative">
                                <button
                                    onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                                    className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors cursor-pointer"
                                >
                                    <ArrowUpDown className="w-4 h-4" />
                                    <span>Sort: {selectedFilter}</span>
                                </button>

                                {isDropdownOpen && (
                                    <div className="absolute right-0 top-full mt-1 w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-10">
                                        {sortOptions.map((option) => (
                                            <button
                                                key={option}
                                                onClick={() => {
                                                    setSelectedFilter(option);
                                                    setIsDropdownOpen(false);
                                                }}
                                                className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 first:rounded-t-lg last:rounded-b-lg cursor-pointer ${
                                                    selectedFilter === option ? 'bg-blue-50 text-blue-600' : 'text-gray-700'
                                                }`}
                                            >
                                                {option}
                                            </button>
                                        ))}
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    {showDomainInfo && (
                        <div className="mb-6 p-4 bg-amber-50 border border-amber-200 rounded-xl flex items-start justify-between gap-3 text-sm text-amber-800 animate-in fade-in duration-150">
                            <div className="flex items-start gap-3">
                                <AlertTriangle className="w-5 h-5 text-amber-600 shrink-0 mt-0.5" />
                                <div>
                                    <p className="font-semibold text-amber-900">
                                        Namespace Collision Detected on Non-Wildcard Domain
                                    </p>
                                    <p className="mt-1 text-xs text-amber-700 leading-relaxed">
                                        Under this non-wildcard domain, all apps mount directly under <code className="bg-amber-100 px-1 py-0.5 rounded font-mono">/zz/space/&lt;namespace&gt;</code> without dedicated subdomains. Only the earliest installed space owns the namespace. Conflicting duplicate spaces have their namespace key striked and &ldquo;Run&rdquo; disabled. To run both, configure and access via a wildcard domain (e.g. <code className="bg-amber-100 px-1 py-0.5 rounded font-mono">*.localhost</code> or <code className="bg-amber-100 px-1 py-0.5 rounded font-mono">*.example.com</code>).
                                    </p>
                                </div>
                            </div>
                            <button
                                type="button"
                                onClick={() => setShowDomainInfo(false)}
                                className="text-amber-500 hover:text-amber-800 p-1 rounded-md transition-colors cursor-pointer shrink-0"
                                title="Close"
                            >
                                <X className="w-4 h-4" />
                            </button>
                        </div>
                    )}

                    {loader.data?.spaces.length === 0 && <EmptySpacesState />}

                    {loader.loading && (<><SimpleLoader /></>)}

                    {/* Empty search/filter state */}
                    {displaySpaces.length === 0 && loader.data && loader.data.spaces.length > 0 && (
                        <div className="text-center py-12 bg-white rounded-xl border border-gray-200 p-8">
                            <p className="text-base text-gray-700 font-medium mb-2">No spaces found matching your current filters.</p>
                            <p className="text-sm text-gray-500 mb-4">
                                {searchTerm
                                    ? `Try clearing your search "${searchTerm}"`
                                    : !showPlugins && !showRootApps
                                    ? 'Try enabling "Show plugins" or "Show root apps"'
                                    : 'Try adjusting your filters'}
                            </p>
                            <div className="flex items-center justify-center gap-3 flex-wrap">
                                {searchTerm && (
                                    <button
                                        onClick={() => setSearchTerm('')}
                                        className="px-3 py-1.5 text-sm bg-gray-100 hover:bg-gray-200 text-gray-700 rounded-lg transition-colors cursor-pointer"
                                    >
                                        Clear Search
                                    </button>
                                )}
                                {!showPlugins && (
                                    <button
                                        onClick={() => changeShowPlugins(true)}
                                        className="px-3 py-1.5 text-sm bg-purple-50 hover:bg-purple-100 text-purple-700 rounded-lg transition-colors cursor-pointer"
                                    >
                                        Show Plugins
                                    </button>
                                )}
                                {!showRootApps && (
                                    <button
                                        onClick={() => changeShowRootApps(true)}
                                        className="px-3 py-1.5 text-sm bg-indigo-50 hover:bg-indigo-100 text-indigo-700 rounded-lg transition-colors cursor-pointer"
                                    >
                                        Show Root Apps
                                    </button>
                                )}
                            </div>
                        </div>
                    )}

                    {/* Cards View */}
                    {viewMode === 'cards' && displaySpaces.length > 0 && (
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                            {displaySpaces.map((space) => {
                                const conflict = spaceConflictMap.get(space.space_id);
                                return (
                                    <SpaceCard
                                        key={space.space_id}
                                        isFavorite={favorites.includes(space.space_id)}
                                        onToggleFavorite={() => {
                                            if (favorites.includes(space.space_id)) {
                                                removeFavorite(space.space_id);
                                            } else {
                                                addFavorite(space.space_id);
                                            }
                                        }}
                                        actionHandler={(action: string) => handleAction(action, space)}
                                        space={{
                                            id: space.space_id,
                                            title: space.package_name,
                                            description: space.package_info,
                                            author: space.package_author,
                                            gradient: space.gradient,
                                            nskey: space.namespace_key,
                                            is_plugin: space.is_plugin,
                                            space_type: space.space_type,
                                            isConflict: conflict?.isConflict,
                                            winnerSpaceId: conflict?.winnerSpaceId,
                                        }}
                                    />
                                );
                            })}
                        </div>
                    )}

                    {/* List View (Grouped in order by package and showing all spaces in that package) */}
                    {viewMode === 'list' && packageGroups.length > 0 && (
                        <div className="space-y-6">
                            {packageGroups.map((group) => {
                                const packageName = group.pkg?.name || group.spaces[0]?.package_name || `Package #${group.installId}`;
                                const packageInfo = group.pkg?.info || group.spaces[0]?.package_info;
                                const packageVersion = group.pkg?.version || group.spaces[0]?.package_version;
                                const packageAuthor = group.pkg?.author_name || group.spaces[0]?.package_author;

                                return (
                                    <div
                                        key={group.installId}
                                        className="bg-white border border-gray-200 rounded-xl shadow-xs overflow-hidden transition-all hover:shadow-sm"
                                    >
                                        {/* Package Header */}
                                        <div className="bg-gray-50/90 border-b border-gray-200 px-6 py-4 flex flex-col md:flex-row md:items-center justify-between gap-3">
                                            <div className="flex items-center gap-3">
                                                <div className="w-10 h-10 rounded-lg bg-blue-100 text-blue-700 flex items-center justify-center font-bold">
                                                    <Box className="w-5 h-5" />
                                                </div>
                                                <div>
                                                    <div className="flex items-center gap-2.5 flex-wrap">
                                                        <h3 className="text-base font-bold text-gray-900">
                                                            {packageName}
                                                        </h3>
                                                        {packageVersion && (
                                                            <span className="bg-gray-200/80 text-gray-700 text-xs px-2 py-0.5 rounded-full font-mono font-medium">
                                                                v{packageVersion}
                                                            </span>
                                                        )}
                                                        <span className="bg-blue-50 text-blue-700 border border-blue-100 text-xs px-2 py-0.5 rounded-full font-medium">
                                                            {group.spaces.length} {group.spaces.length === 1 ? 'space' : 'spaces'}
                                                        </span>
                                                    </div>
                                                    {packageInfo && (
                                                        <p className="text-xs text-gray-500 mt-0.5 line-clamp-1">
                                                            {packageInfo}
                                                        </p>
                                                    )}
                                                </div>
                                            </div>

                                            {packageAuthor && (
                                                <div className="text-xs text-gray-500 flex items-center gap-1.5 self-start md:self-auto">
                                                    <Users className="w-3.5 h-3.5 text-gray-400" />
                                                    <span>{packageAuthor}</span>
                                                </div>
                                            )}
                                        </div>

                                        {/* Spaces in this package */}
                                        <div className="divide-y divide-gray-100">
                                            {group.spaces.map((space) => {
                                                const isFav = favorites.includes(space.space_id);
                                                const conflict = spaceConflictMap.get(space.space_id);
                                                const isConflict = Boolean(conflict?.isConflict);
                                                const winnerSpaceId = conflict?.winnerSpaceId;
                                                return (
                                                    <div
                                                        key={space.space_id}
                                                        className="px-6 py-3.5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 hover:bg-gray-50/60 transition-colors"
                                                    >
                                                        <div className="flex items-center gap-3 flex-wrap">
                                                            <div className={`w-8 h-8 rounded-lg bg-gradient-to-br ${space.gradient} flex items-center justify-center text-white text-xs font-bold shadow-xs`}>
                                                                #{space.space_id}
                                                            </div>
                                                            <div className="flex items-center gap-2 flex-wrap">
                                                                <span
                                                                    className={`font-mono text-sm font-semibold px-2.5 py-1 rounded-md transition-colors ${
                                                                        isConflict
                                                                            ? 'line-through text-red-600 bg-red-50 decoration-red-500 decoration-2'
                                                                            : 'text-gray-900 bg-gray-100'
                                                                    }`}
                                                                    title={
                                                                        isConflict
                                                                            ? `Namespace collision: taken by space #${winnerSpaceId} on non-wildcard domain`
                                                                            : undefined
                                                                    }
                                                                >
                                                                    {space.namespace_key}
                                                                </span>
                                                                {isConflict && (
                                                                    <span
                                                                        className="bg-red-100 text-red-700 text-xs px-2 py-0.5 rounded-full font-medium border border-red-200"
                                                                        title={`Namespace is taken by Space #${winnerSpaceId} on this non-wildcard domain`}
                                                                    >
                                                                        Conflict
                                                                    </span>
                                                                )}
                                                                {space.is_plugin ? (
                                                                    <span className="bg-purple-100 text-purple-700 text-xs px-2 py-0.5 rounded-full flex items-center gap-1 font-medium border border-purple-200">
                                                                        <Plug className="w-3 h-3" /> Plugin
                                                                    </span>
                                                                ) : space.space_type === 'RootApp' ? (
                                                                    <span className="bg-indigo-100 text-indigo-700 text-xs px-2 py-0.5 rounded-full font-medium border border-indigo-200 flex items-center gap-1">
                                                                        <Globe className="w-3 h-3" /> RootApp
                                                                    </span>
                                                                ) : (
                                                                    <span className="bg-emerald-50 text-emerald-700 text-xs px-2 py-0.5 rounded-full font-medium border border-emerald-200">
                                                                        App
                                                                    </span>
                                                                )}
                                                            </div>
                                                        </div>

                                                        <div className="flex items-center gap-2 self-end sm:self-auto">
                                                            <button
                                                                type="button"
                                                                onClick={() => {
                                                                    if (isFav) {
                                                                        removeFavorite(space.space_id);
                                                                    } else {
                                                                        addFavorite(space.space_id);
                                                                    }
                                                                }}
                                                                className="p-2 text-gray-400 hover:text-red-500 transition-colors cursor-pointer rounded-lg hover:bg-gray-100"
                                                                title={isFav ? "Remove from favorites" : "Add to favorites"}
                                                            >
                                                                <Heart className={`w-4 h-4 ${isFav ? 'fill-red-500 text-red-500' : ''}`} />
                                                            </button>

                                                            <button
                                                                type="button"
                                                                disabled={isConflict}
                                                                onClick={() => {
                                                                    if (isConflict) return;
                                                                    router.push(`/portal/admin/exec?nskey=${space.namespace_key}&space_id=${space.space_id}`);
                                                                }}
                                                                className={`flex items-center gap-1.5 text-xs font-medium px-3 py-1.5 rounded-lg transition-colors ${
                                                                    isConflict
                                                                        ? 'bg-gray-100 text-gray-400 cursor-not-allowed opacity-50'
                                                                        : 'bg-blue-50 text-blue-600 hover:bg-blue-100 cursor-pointer'
                                                                }`}
                                                                title={
                                                                    isConflict
                                                                        ? `Run disabled: namespace taken by space #${winnerSpaceId} on non-wildcard domain`
                                                                        : undefined
                                                                }
                                                            >
                                                                <CloudLightning className="w-3.5 h-3.5" />
                                                                <span>Run</span>
                                                            </button>

                                                            <button
                                                                type="button"
                                                                onClick={() => {
                                                                    const params = new URLSearchParams();
                                                                    params.set("install_id", space.install_id.toString());
                                                                    params.set("space_id", space.space_id.toString());
                                                                    params.set("namespace_key", space.namespace_key);
                                                                    params.set("nskey", space.namespace_key);
                                                                    params.set("package_version_id", space.package_version_id.toString());
                                                                    router.push(`/portal/admin/spaces/tools/overview?${params.toString()}`);
                                                                }}
                                                                className="flex items-center gap-1 text-xs font-medium px-2.5 py-1.5 border border-gray-200 text-gray-700 rounded-lg hover:bg-gray-100 transition-colors cursor-pointer"
                                                                title="Tools & Settings"
                                                            >
                                                                <Wrench className="w-3.5 h-3.5" />
                                                                <span className="hidden md:inline">Tools</span>
                                                            </button>

                                                            <ActionDropdown
                                                                disableRun={isConflict}
                                                                onClick={(action) => handleAction(action, space)}
                                                            />
                                                        </div>
                                                    </div>
                                                );
                                            })}
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}
                </div>
            </div>
        </WithAdminBodyLayout>
    );
};

const SpaceCard = ({ space, actionHandler, isFavorite, onToggleFavorite }: { space: any, actionHandler: any, isFavorite: boolean, onToggleFavorite: () => void }) => {
    const router = useRouter();

    return (
        <div className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${space.gradient} p-6 text-white min-h-[200px] group hover:scale-105 transition-transform duration-200 `}>
            <div className="flex flex-col h-full justify-between">
                <div>
                    <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2 text-sm flex-wrap">
                            <span className="font-semibold">
                                #{space.id}
                            </span>
                            <span
                                className={`bg-white/20 backdrop-blur-sm px-2 py-1 rounded text-sm font-mono ${
                                    space.isConflict ? 'line-through text-red-200 decoration-red-400 decoration-2' : ''
                                }`}
                                title={
                                    space.isConflict
                                        ? `Namespace collision: taken by space #${space.winnerSpaceId} on non-wildcard domain`
                                        : undefined
                                }
                            >
                                {space.nskey}
                            </span>
                            {space.isConflict && (
                                <span
                                    className="bg-red-500/80 text-white px-2 py-0.5 rounded text-xs font-medium"
                                    title={`Namespace conflict: taken by space #${space.winnerSpaceId} on this non-wildcard domain`}
                                >
                                    Conflict
                                </span>
                            )}
                            {space.is_plugin && (
                                <span className="bg-purple-600/80 px-2 py-1 rounded text-xs font-medium flex items-center gap-1">
                                    <Plug className="w-3 h-3" /> Plugin
                                </span>
                            )}
                            {space.space_type === 'RootApp' && (
                                <span className="bg-indigo-600/80 px-2 py-1 rounded text-xs font-medium flex items-center gap-1">
                                    <Globe className="w-3 h-3" /> RootApp
                                </span>
                            )}
                            {space.mcp && (
                                <span className="bg-pink-500/80 px-2 py-1 rounded text-xs">🔥 MCP</span>
                            )}
                        </div>

                        <div className='flex justify-end'>
                            <button
                                onClick={(e) => {
                                    e.stopPropagation();
                                    onToggleFavorite();
                                }}
                                className="flex items-center gap-1 text-xs hover:scale-110 transition-transform cursor-pointer"
                            >
                                <Heart
                                    className={`w-4 h-4 ${isFavorite ? 'fill-red-500 text-red-500' : 'text-white/80'}`}
                                />
                            </button>
                        </div>
                    </div>

                    <h3 className="text-xl font-bold mb-2">{space.title}</h3>
                    <p className="text-sm text-white/90 mb-4 line-clamp-2">{space.description}</p>
                </div>

                <div className="flex items-center justify-between text-sm">
                    <div className="flex items-center gap-2">
                        <div className="w-6 h-6 bg-white/20 rounded-full flex items-center justify-center">
                            <Users className="w-3 h-3" />
                        </div>

                        <div className="flex flex-col">
                            <span className="font-medium">{space.author}</span>
                            <div className="flex items-center gap-1 text-xs">
                                <span>{space.timeAgo}</span>
                            </div>
                        </div>
                    </div>

                    <div className="flex gap-2">
                        <button
                            disabled={space.isConflict}
                            className={`flex items-center gap-1 text-xs px-3 py-2 rounded-lg transition-colors ${
                                space.isConflict
                                    ? 'bg-white/10 text-white/40 cursor-not-allowed opacity-50'
                                    : 'bg-white/20 backdrop-blur-sm hover:bg-white/40 cursor-pointer hover:text-blue-600'
                            }`}
                            onClick={() => {
                                if (space.isConflict) return;
                                router.push(`/portal/admin/exec?nskey=${space.nskey}&space_id=${space.id}`);
                            }}
                            title={
                                space.isConflict
                                    ? `Run disabled: namespace conflict on non-wildcard domain (owned by #${space.winnerSpaceId})`
                                    : undefined
                            }
                        >
                            <CloudLightning className="w-4 h-4" />
                            <span>Run</span>
                        </button>

                        <ActionDropdown
                            disableRun={space.isConflict}
                            onClick={actionHandler}
                        />
                    </div>
                </div>
            </div>
        </div>
    );
};


const actionsOptions = [
    { id: "run", label: "Run in new tab", icon: <Bolt className="w-4 h-4" /> },
    { id: "tools", label: "Tools and settings", icon: <Box className="w-4 h-4" /> },
    { id: "delete", label: "Delete", icon: <Trash2Icon className="w-4 h-4" /> }
]


interface ActionDropdownProps {
    onClick: (action: string) => void;
    disableRun?: boolean;
}

const ActionDropdown = (props: ActionDropdownProps) => {
    const [isDropdownOpen, setIsDropdownOpen] = useState(false);
    const [buttonRect, setButtonRect] = useState<DOMRect | null>(null);
    const buttonRef = useRef<HTMLButtonElement>(null);

    const handleToggleDropdown = () => {
        if (!isDropdownOpen && buttonRef.current) {
            const rect = buttonRef.current.getBoundingClientRect();
            setButtonRect(rect);
        }
        setIsDropdownOpen(!isDropdownOpen);
    };

    useEffect(() => {
        const dropdownRef = document.getElementById("action-dropdown");
        const handleClickOutside = (event: MouseEvent) => {
            if (
                isDropdownOpen &&
                buttonRef.current &&
                !buttonRef.current.contains(event.target as Node) &&
                dropdownRef &&
                !dropdownRef.contains(event.target as Node)
            ) {
                setIsDropdownOpen(false);
            }
        };

        const handleScroll = () => {
            if (isDropdownOpen && buttonRef.current) {
                const rect = buttonRef.current.getBoundingClientRect();
                setButtonRect(rect);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        window.addEventListener('scroll', handleScroll, true);
        window.addEventListener('resize', handleScroll);

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
            window.removeEventListener('scroll', handleScroll, true);
            window.removeEventListener('resize', handleScroll);
        };
    }, [isDropdownOpen]);

    return (
        <>
            <div className="flex items-center gap-4">
                <div className="relative">
                    <button
                        ref={buttonRef}
                        onClick={handleToggleDropdown}
                        className="flex items-center gap-2 px-3 py-2 border border-gray-300 rounded-lg text-sm hover:bg-gray-50 transition-colors hover:text-blue-600 cursor-pointer"
                    >
                        <Bolt className="w-4 h-4" />
                        <span>Actions</span>
                    </button>
                </div>
            </div>

            {/* Render dropdown in a portal */}
            {isDropdownOpen && buttonRect && createPortal(
                <div
                    id="action-dropdown"
                    className="fixed w-48 bg-white border border-gray-300 rounded-lg shadow-lg z-[9999]"
                    style={{
                        top: buttonRect.bottom + 4,
                        left: buttonRect.right - 192,
                    }}
                >
                    {actionsOptions.map((option) => {
                        const isRunDisabled = option.id === "run" && props.disableRun;
                        return (
                            <button
                                key={option.id}
                                disabled={isRunDisabled}
                                onClick={async () => {
                                    if (isRunDisabled) return;
                                    console.log("clicked", option.id);
                                    props.onClick(option.id);
                                    setTimeout(() => {
                                        setIsDropdownOpen(false);
                                    }, 100);
                                }}
                                className={`w-full text-left px-3 py-2 text-sm first:rounded-t-lg last:rounded-b-lg transition-colors ${
                                    isRunDisabled
                                        ? 'text-gray-300 cursor-not-allowed line-through'
                                        : 'text-gray-700 hover:text-blue-600 hover:bg-gray-200 cursor-pointer'
                                }`}
                                title={isRunDisabled ? "Run disabled: namespace conflict on non-wildcard domain" : undefined}
                            >
                                <div className="inline-flex items-center gap-2">
                                    {option.icon}
                                    {option.label}
                                </div>
                            </button>
                        );
                    })}
                </div>,
                document.body
            )}
        </>
    );
};

const EmptySpacesState = () => {
    const router = useRouter();

    return (
        <div className="flex flex-col items-center justify-center py-16 px-6">
            <div className="mb-8">
                <svg
                    width="200"
                    height="160"
                    viewBox="0 0 200 160"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="drop-shadow-sm"
                >
                    {/* Decorative circles */}
                    <circle cx="50" cy="40" r="3" fill="#e5e7eb" opacity="0.5" />
                    <circle cx="150" cy="30" r="2" fill="#e5e7eb" opacity="0.3" />
                    <circle cx="170" cy="60" r="2.5" fill="#e5e7eb" opacity="0.4" />

                    {/* Box/package icon */}
                    <rect x="70" y="50" width="60" height="60" rx="4" fill="#f3f4f6" stroke="#e5e7eb" strokeWidth="2" />
                    <rect x="80" y="60" width="40" height="30" rx="2" fill="#e5e7eb" opacity="0.5" />
                    <line x1="100" y1="50" x2="100" y2="110" stroke="#e5e7eb" strokeWidth="2" />
                    <line x1="70" y1="80" x2="130" y2="80" stroke="#e5e7eb" strokeWidth="2" />

                    {/* Decorative stars */}
                    <g className="animate-bounce" style={{ animationDelay: '0.5s' }}>
                        <path d="M65 25l2 6 6 2-6 2-2 6-2-6-6-2 6-2 2-6z" fill="#fbbf24" />
                    </g>
                    <g className="animate-bounce" style={{ animationDelay: '1s' }}>
                        <path d="M140 20l1.5 4.5 4.5 1.5-4.5 1.5-1.5 4.5-1.5-4.5-4.5-1.5 4.5-1.5 1.5-4.5z" fill="#f59e0b" />
                    </g>
                    <g className="animate-bounce" style={{ animationDelay: '1.5s' }}>
                        <path d="M30 80l1 3 3 1-3 1-1 3-1-3-3-1 3-1 1-3z" fill="#fbbf24" />
                    </g>
                </svg>
            </div>

            {/* Content */}
            <div className="text-center max-w-md">
                <h3 className="text-xl font-semibold text-gray-700 mb-3">
                    No spaces installed yet!
                </h3>
                <p className="text-sm text-gray-600 mb-6">
                    Get started by installing your first space from the store.
                </p>

                <div className="flex justify-center">
                    <button
                        className="bg-gradient-to-r from-blue-500 to-purple-600 text-white px-6 py-3 rounded-lg font-medium hover:from-blue-600 hover:to-purple-700 transition-all transform hover:scale-105 shadow-lg"
                        onClick={() => router.push("/portal/admin/store")}
                    >
                        Go to Store
                    </button>
                </div>
            </div>
        </div>
    );
};



