import { Search } from "lucide-react";

interface PropsType {
    searchText?: string;
    setSearchText: (text: string) => void;
}


const BigSearchBar: React.FC<PropsType> = (props: PropsType) => {

    return (
        <div className="relative w-full max-w-2xl mx-auto">
            <div className="absolute inset-0 bg-gradient-to-r from-purple-300 to-pink-300 opacity-50 rounded-lg"></div>
            <div className="relative z-10 p-6 bg-white rounded-lg shadow-lg">
                <div className="flex items-center">
                    <Search className="text-gray-400 w-5 h-5 mr-3" />
                    <input
                        type="text"
                        placeholder="Search spaces..."
                        className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        value={props.searchText}
                        onChange={(e) => props.setSearchText(e.target.value)}
                    />
                </div>
            </div>
        </div>
    );
}

export default BigSearchBar;