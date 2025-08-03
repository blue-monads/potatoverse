


interface PropsType {
    name: string;
    description: string;
    Icon: React.ElementType;
    children: React.ReactNode;
    rightContent?: React.ReactNode;
}

const WithAdminBodyLayout = (props: PropsType) => {
    return (
        <div className="flex flex-col min-h-screen bg-gray-100">
            <header className="bg-white border-b border-gray-200 px-6 py-4">
                <div className="max-w-7xl mx-auto flex items-center justify-between">
                    <div className="flex items-center gap-4">
                        <div className="flex items-center gap-2">
                            <div className="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg flex items-center justify-center">
                                <props.Icon className="w-5 h-5 text-white" />
                            </div>
                            <div>
                                <h1 className="text-xl font-bold">{props.name}</h1>
                                <p className="text-sm text-gray-600">{props.description}</p>
                            </div>
                        </div>
                    </div>

                    <div className="flex items-center gap-4">
                        {props.rightContent}
                    </div>
                </div>
            </header>

            {props.children}
        </div>
    );

}


export default WithAdminBodyLayout;