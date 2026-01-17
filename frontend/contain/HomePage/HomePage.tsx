

"use client";
import React, { useEffect, useState } from 'react';
import { Heart, Users, SquareUserRound, Cog, Link2Icon } from 'lucide-react';
import { Clock, ArrowRight, Code } from 'lucide-react';
import EmptyFavorite from './sub/EmptyFavorite';
import HeroSection from './sub/HeroSection';
import useSimpleDataLoader from '@/hooks/useSimpleDataLoader';
import { AdminPortalData, getAdminPortalData, InstalledSpace, listInstalledSpaces } from '@/lib/api';
import { useGApp } from '@/hooks';
import { useRouter } from 'next/navigation';
import useFavorites from '@/hooks/useFavorites/useFavorites';
import { formatSpace, FormattedSpace } from '@/app/portal/admin/spaces/page';


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
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">


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
                    onClick={() => router.push("/portal/admin/setting")}
                    >

                        <Cog className="w-8 h-8 text-yellow-600" />
                        <div>
                            <h3 className="text-lg font-semibold">Settings</h3>
                            <p className="text-sm text-gray-600">Configure application settings</p>
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