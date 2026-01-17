import { useEffect, useState } from "react";


const useFavorites = () => {
    const [favorites, setFavorites] = useState<number[]>([]);

    const addFavorite = (spaceId: number) => {
        setFavorites([...favorites, spaceId]);
    }

    const removeFavorite = (spaceId: number) => {
        setFavorites(favorites.filter((id) => id !== spaceId));
    }

    const loadFavorites = () => {
        const fromls = JSON.parse(localStorage.getItem('__favorite_space_ids__') || '[]');
        setFavorites(fromls as number[]);
    }

    const saveFavorites = () => {
        const tols = JSON.stringify(favorites);
        localStorage.setItem('__favorite_space_ids__', tols);
    }

    useEffect(() => {

        try {
            loadFavorites();
        } catch (error) {
            console.error('Error parsing favorites:', error);
        }



    }, []);



    useEffect(() => {

        const existing = localStorage.getItem('__favorite_space_ids__') || '[]';
        const existingFavorites = JSON.parse(existing) as number[] || [];

        if (existingFavorites.length !== favorites.length) {
            saveFavorites();
            return;
        }

        for (const favorite of favorites) {
            if (!existingFavorites.includes(favorite)) {
                saveFavorites();
                return;
            }
        }

        for (const favorite of existingFavorites) {
            if (!favorites.includes(favorite)) {
                saveFavorites();
                return;
            }
        }


    }, [favorites]);



    return { favorites, addFavorite, removeFavorite };
}

export default useFavorites;