import { Search, Sparkles, Zap } from "lucide-react";


interface PropsType {
    searchTerm: string;
    setSearchTerm: (value: string) => void;
    popularSearches?: string[];
}


const HeroSection = (props: PropsType) => {
    return (<>


        <div className="bg-gradient-to-br from-blue-600 via-purple-600 to-pink-600 text-white w-full flex items-center justify-center">
            <div className="max-w-7xl mx-auto px-6 py-16">
                <div className="text-center max-w-4xl mx-auto">
                    <div className="flex items-center justify-center gap-2 mb-4">
                        <Sparkles className="w-6 h-6" />
                        <span className="text-lg font-medium">Welcome to admin portal!</span>
                    </div>
                    <h1 className="text-5xl font-bold mb-6">
                        Discover Apps and Tools
                    </h1>
                    <p className="text-xl text-white/90 mb-8 leading-relaxed">
                        Find apps to fit your needs or create your own quickly.
                    </p>

                    {/* Search Bar */}
                    <div className="relative max-w-2xl mx-auto">
                        <Search className="absolute left-4 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                        <input
                            type="text"
                            placeholder="Search recent apps"
                            className="w-full pl-12 pr-16 py-4 text-gray-900 bg-white rounded-xl focus:outline-none focus:ring-4 focus:ring-white/30 shadow-lg text-lg"
                            value={props.searchTerm}
                            onChange={(e) => props.setSearchTerm(e.target.value)}
                        />
                        <button className="absolute right-2 top-1/2 transform -translate-y-1/2 bg-blue-600 text-white p-2 rounded-lg hover:bg-blue-700 transition-colors">
                            <Zap className="w-5 h-5" />
                        </button>
                    </div>

                    <div className="flex items-center justify-center gap-4 mt-6 text-sm text-white/80">
                        <span>Popular searches:</span>
                        {props.popularSearches?.map((search) => (
                            <button key={search}
                                className="bg-white/20 px-3 py-1 rounded-full hover:bg-white/30 transition-colors"
                                onClick={() => props.setSearchTerm(search)}
                            >
                                {search}
                            </button>
                        ))}

                        {!props.popularSearches && (

                            <>

                                <div className="flex gap-2">
                                    <div className="bg-white/20 px-3 py-1 rounded-full hover:bg-white/30 transition-colors h-6 w-16 animate-pulse" />

                                    <div className="bg-white/20 px-3 py-1 rounded-full hover:bg-white/30 transition-colors h-6 w-16 animate-pulse" />

                                </div>
                            </>

                        )}



                    </div>
                </div>
            </div>
        </div>


    </>)
}


export default HeroSection;