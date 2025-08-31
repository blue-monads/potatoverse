const EmptyFavorite = () => {
    return (
        <div className="flex flex-col items-center justify-center max-w-7xl mx-auto px-6 py-4">
            <div className="mb-8">
                <svg
                    width="200"
                    height="160"
                    viewBox="0 0 200 160"
                    fill="none"
                    xmlns="http://www.w3.org/2000/svg"
                    className="drop-shadow-sm"
                >

                    <circle cx="50" cy="40" r="3" fill="#e5e7eb" opacity="0.5" />
                    <circle cx="150" cy="30" r="2" fill="#e5e7eb" opacity="0.3" />
                    <circle cx="170" cy="60" r="2.5" fill="#e5e7eb" opacity="0.4" />

                    <path
                        d="M100 130c-2-1.5-45-35-45-65 0-20 15-35 35-35 10 0 18 5 23 12 5-7 13-12 23-12 20 0 35 15 35 35 0 30-43 63.5-45 65z"

                        className="animate-pulse"
                    />

                    <path
                        d="M100 130c-2-1.5-45-35-45-65 0-20 15-35 35-35 10 0 18 5 23 12 5-7 13-12 23-12 20 0 35 15 35 35 0 30-43 63.5-45 65z"
                        stroke="#f3f4f6"
                        strokeWidth="3"
                        fill="none"
                    />

                    <g className="animate-bounce" style={{ animationDelay: '0.5s' }}>
                        <path d="M65 25l2 6 6 2-6 2-2 6-2-6-6-2 6-2 2-6z" fill="#fbbf24" />
                    </g>
                    <g className="animate-bounce" style={{ animationDelay: '1s' }}>
                        <path d="M140 20l1.5 4.5 4.5 1.5-4.5 1.5-1.5 4.5-1.5-4.5-4.5-1.5 4.5-1.5 1.5-4.5z" fill="#f59e0b" />
                    </g>
                    <g className="animate-bounce" style={{ animationDelay: '1.5s' }}>
                        <path d="M30 80l1 3 3 1-3 1-1 3-1-3-3-1 3-1 1-3z" fill="#fbbf24" />
                    </g>

                    <defs>
                        <linearGradient id="heartGradient" x1="0%" y1="0%" x2="100%" y2="100%">
                            <stop offset="0%" stopColor="#fce7f3" />
                            <stop offset="50%" stopColor="#f9a8d4" />
                            <stop offset="100%" stopColor="#ec4899" />
                        </linearGradient>
                    </defs>
                </svg>
            </div>

            {/* Content */}
            <div className="text-center max-w-md">


                <h3 className="text-xl font-semibold text-gray-700 mb-3">
                    No favorites yet!
                </h3>


                <div className="flex flex-col sm:flex-row gap-3 justify-center">
                    <button className="bg-gradient-to-r from-pink-500 to-purple-600 text-white px-6 py-3 rounded-lg font-medium hover:from-pink-600 hover:to-purple-700 transition-all transform hover:scale-105 shadow-lg">
                        Explore Apps
                    </button>
                    <button className="border border-gray-300 text-gray-700 px-6 py-3 rounded-lg font-medium hover:bg-gray-50 transition-colors">
                        Store
                    </button>
                </div>
            </div>

            {/* Bottom decoration */}
            <div className="mt-12 flex items-center gap-2 text-sm text-gray-400">
                <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M13 3c-4.97 0-9 4.03-9 9H1l3.89 3.89.07.14L9 12H6c0-3.87 3.13-7 7-7s7 3.13 7 7-3.13 7-7 7c-1.93 0-3.68-.79-4.94-2.06l-1.42 1.42C8.27 19.99 10.51 21 13 21c4.97 0 9-4.03 9-9s-4.03-9-9-9zm-1 5v5l4.28 2.54.72-1.21-3.5-2.08V8H12z" />
                </svg>
                <span>Start adding favorites to see them here</span>
            </div>
        </div>
    );
};

export default EmptyFavorite;