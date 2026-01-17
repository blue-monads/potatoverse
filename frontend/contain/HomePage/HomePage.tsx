

"use client";
import React, { useEffect, useState } from 'react';
import { Heart, Users, SquareUserRound, Cog, Link2Icon, StoreIcon } from 'lucide-react';
import { Clock, ArrowRight, Code } from 'lucide-react';
import EmptyFavorite from './sub/EmptyFavorite';
import HeroSection from './sub/HeroSection';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { AdminPortalData, getAdminPortalData, InstalledSpace, listInstalledSpaces } from '@/lib/api';
import { useGApp } from '@/hooks';
import { useRouter } from 'next/navigation';
import useFavorites from '@/hooks/useFavorites/useFavorites';
import { formatSpace, FormattedSpace } from '@/lib';


export default function HomePage() {

    const router = useRouter();
    const [searchTerm, setSearchTerm] = useState('');
    const gapp = useGApp();

    const loader = useSimpleDataLoader<AdminPortalData>({
        loader: () => getAdminPortalData("admin"),
        ready: gapp.isInitialized,
    });


    const [favSpaces, setFavSpaces] = useState<FormattedSpace[]>([]);
    const favorites = useFavorites();


    const load = async () => {

        try {

            const resp = await listInstalledSpaces();
            if (resp.status !== 200) {
                return;
            }

            const nextFormattedSpaces = formatSpace(resp.data);

            const nextfavs = nextFormattedSpaces.filter((space) => favorites.favorites.includes(space.space_id));
            setFavSpaces(nextfavs);

        } catch (error) {
            console.error(error);
        }
    }


    useEffect(() => {
        if (!favorites.favoritesLoaded) return;

        load();

    }, [favorites.favorites, favorites.favoritesLoaded]);





    // favs.length = 0; // For testing empty state


    return (<>
        <div className="min-h-screen bg-gray-50 w-full">
            {/* Hero Section */}

            <HeroSection
                searchTerm={searchTerm}
                setSearchTerm={setSearchTerm}
                popularSearches={loader.data?.popular_keywords || []}
            />



            {/* Favorites Section */}

            <div className="max-w-7xl mx-auto px-6 py-12">
                <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                        <Heart className="w-6 h-6 text-pink-600" />
                        <h2 className="text-2xl font-bold text-gray-900">Favorites</h2>
                    </div>
                </div>

                {favSpaces.length === 0 ? (<>
                    <EmptyFavorite />
                </>) : (<>

                    <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 w-full">
                        <div className="lg:col-span-2">

                            <div className="grid grid-cols-1 md:grid-cols-2 gap-4 w-full">

                                {favSpaces.map((space) => (
                                    <FavCard
                                        key={space.space_id}
                                        app={space}

                                        onClick={() => router.push(`/portal/admin/exec?nskey=${space.namespace_key}&space_id=${space.space_id}`)}

                                    />
                                ))}

                            </div>


                        </div>


                    </div>

                </>)}
            </div>




            {/* Quick Links (users, profile, setting, dev console) */}
            <div className="max-w-7xl mx-auto px-6 py-12">

                <div className="flex items-center justify-between mb-6">
                    <div className="flex items-center gap-3">
                        <Link2Icon className="w-6 h-6 text-pink-600" />
                        <h2 className="text-2xl font-bold text-gray-900">Quick Links</h2>
                    </div>
                </div>
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 cursor-pointer">


                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200"
                        onClick={() => router.push("/portal/admin/users")}

                    >
                        <Users className="w-8 h-8 text-blue-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Users</h3>
                            <p className="text-sm text-gray-600">Manage users and permissions</p>
                        </div>
                    </div>

                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200"
                        onClick={() => router.push("/portal/admin/profile")}
                    >
                        <SquareUserRound className="w-8 h-8 text-green-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Profile</h3>
                            <p className="text-sm text-gray-600">View and edit your profile</p>
                        </div>
                    </div>

                    <div className="bg-white rounded-lg shadow-lg p-6 flex items-center gap-4 hover:shadow-xl transition-shadow border border-gray-200"
                        onClick={() => router.push("/portal/admin/store")}
                    >

                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            fill="none"
                            viewBox="0 0 24 24"
                            strokeWidth={1.5}
                            stroke="currentColor"
                            className="w-8 h-8 text-yellow-600"
                        >
                            <path
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                d="M14.25 6.087c0-.355.186-.676.401-.959.221-.29.349-.634.349-1.003 0-1.036-1.007-1.875-2.25-1.875s-2.25.84-2.25 1.875c0 .369.128.713.349 1.003.215.283.401.604.401.959v0a.64.64 0 01-.657.643 48.39 48.39 0 01-4.163-.3c.186 1.613.293 3.25.315 4.907a.656.656 0 01-.658.663v0c-.355 0-.676-.186-.959-.401a1.647 1.647 0 00-1.003-.349c-1.036 0-1.875 1.007-1.875 2.25s.84 2.25 1.875 2.25c.369 0 .713-.128 1.003-.349.283-.215.604-.401.959-.401v0c.31 0 .555.26.532.57a48.039 48.039 0 01-.642 5.056c1.518.19 3.058.309 4.616.354a.64.64 0 00.657-.643v0c0-.355-.186-.676-.401-.959a1.647 1.647 0 01-.349-1.003c0-1.035 1.008-1.875 2.25-1.875 1.243 0 2.25.84 2.25 1.875 0 .369-.128.713-.349 1.003-.215.283-.4.604-.4.959v0c0 .333.277.599.61.58a48.1 48.1 0 005.427-.63 48.05 48.05 0 00.582-4.717.532.532 0 00-.533-.57v0c-.355 0-.676.186-.959.401-.29.221-.634.349-1.003.349-1.035 0-1.875-1.007-1.875-2.25s.84-2.25 1.875-2.25c.37 0 .713.128 1.003.349.283.215.604.401.96.401v0a.656.656 0 00.658-.663 48.422 48.422 0 00-.37-5.36c-1.886.342-3.81.574-5.766.689a.578.578 0 01-.61-.58v0z"
                            />
                        </svg>
                        <div>
                            <h3 className="text-lg font-semibold">Store</h3>
                            <p className="text-sm text-gray-600">Browse and install applications</p>
                        </div>
                    </div>

                </div>
            </div>
        </div>








    </>)
}











const FavCard = ({ app, onClick }: { app: FormattedSpace, onClick: () => void }) => (
    <div
        onClick={onClick}
        className={`relative overflow-hidden rounded-xl bg-gradient-to-br ${app.gradient} p-5 text-white hover:scale-105 transition-all duration-200 cursor-pointer group`}>
        <div className="flex flex-col h-full justify-between">
            <div>
                <h3 className="text-lg font-bold mb-2">{app.package_name}</h3>
                <p className="text-sm text-white/90 mb-3 line-clamp-2">{app.package_info}</p>
                <span className="text-xs bg-white/20 px-2 py-1 rounded-full">{app.package_version}</span>
            </div>

            <div className="flex items-center justify-between mt-4 text-sm">
                <div className="flex items-center gap-2">
                    <div className="w-5 h-5 bg-white/20 rounded-full flex items-center justify-center">
                        <Users className="w-3 h-3" />
                    </div>
                    <span className="font-medium">{app.package_author}</span>
                </div>
                {/* <div className="flex items-center gap-1 text-white/80">
                    <Clock className="w-3 h-3" />
                    <span>{app.lastUsed}</span>
                </div> */}
            </div>
        </div>
        <div className="absolute top-3 right-3 opacity-0 group-hover:opacity-100 transition-opacity">
            <ArrowRight className="w-5 h-5" />
        </div>
    </div>
);