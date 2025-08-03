import { Search, Zap } from "lucide-react";

interface PropsType {
    searchText?: string;
    setSearchText: (text: string) => void;
}

const BigSearchBar: React.FC<PropsType> = (props: PropsType) => {

    return (
        <div className="bg-white border-b border-gray-200 px-6 py-4">
            <div className="max-w-7xl mx-auto">
                <div className="relative">
                    <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-5 h-5" />
                    <input
                        type="text"
                        placeholder="Search spaces..."
                        className="w-full pl-10 pr-4 py-3 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        value={props.searchText}
                        onChange={(e) => props.setSearchText(e.target.value)}
                    />
                    <button className="absolute right-3 top-1/2 transform -translate-y-1/2 p-1 cursor-pointer hover:bg-gray-100 rounded-full transition-colors">
                        <Zap className="w-5 h-5 text-gray-400" />
                    </button>
                </div>
            </div>
        </div>
    );
}

export default BigSearchBar;