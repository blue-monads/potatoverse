import Image from "next/image";


interface PropsType {
    children: React.ReactNode;
}

const WithLoginLayout = (props: PropsType) => {
    return (
        <>

            <main className="w-full flex">

                <div className="relative flex-1 hidden items-center justify-center h-screen from-primary-700-300 to-secondary-700-300 bg-gradient-to-b lg:flex">
                    <div className="relative z-10 w-full max-w-md">
                        <Image
                            src="/z/pages/logo.png"
                            alt="Turnix Logo"
                            width={200}
                            height={200}
                        />
                        <div className=" mt-16 space-y-3">
                            <h3 className="text-white text-3xl font-bold">Turnix</h3>
                            <p className="text-gray-300">
                                Platfrom for your apps.
                            </p>

                        </div>
                    </div>
                    <div
                        className="absolute inset-0 my-auto h-[500px] "

                    >

                    </div>
                </div>

                <div className="flex-1 flex items-center justify-center h-screen">
                    {props.children}
                </div>



            </main>



        </>
    )
}

export default WithLoginLayout;