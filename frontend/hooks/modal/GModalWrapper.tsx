import { useGApp } from "../contexts/GAppStateContext";

const GModalWrapper = () => {
    const { modal } = useGApp();
    const { isOpen, modalContent, closeModal } = modal;

    if (!isOpen || !modalContent) {
        return null;
    }


    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
            <div 
                className="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
                onClick={closeModal}
            />
            
            <div className={`relative bg-white rounded-lg shadow-2xl ${getSizeClasses(modalContent.size)} max-h-[90vh] overflow-hidden`}>

                <CloseButton onClose={closeModal} />

                 {modalContent.title && (
                    <div className="flex items-center justify-between p-6 border-b border-gray-200">
                        <h2 className="text-xl font-semibold text-gray-900">
                            {modalContent.title}
                        </h2>

                    </div>
                )}
                
                {/* Content */}
                <div className="p-6">
                    {modalContent.content}
                </div>
                
                {/* Close button for modals without title */}
                {!modalContent.title && (
                    <CloseButton onClose={closeModal} />
                )}
            </div>
        </div>
    );
};

interface CloseButtonProps {
    onClose: () => void;
}

const CloseButton = ({ onClose }: CloseButtonProps) => {
    return (
        <button
            className="text-gray-400 hover:text-gray-600 absolute top-4 right-4 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 rounded-full p-1 transition-colors"
            aria-label="close"
            onClick={onClose}
        >
            <svg
                width={20}
                height={20}
                viewBox="0 0 24 24"
                fill="none"
                xmlns="http://www.w3.org/2000/svg"
            >
                <path
                    d="M18 6L6 18"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                />
                <path
                    d="M6 6L18 18"
                    stroke="currentColor"
                    strokeWidth="2"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                />
            </svg>
        </button>
    );
};

const getSizeClasses = (size?: string) => {
    switch (size) {
        case "sm":
            return "w-80 md:w-96";
        case "md":
            return "w-96 md:w-auto";
        case "lg":
            return "w-full max-w-2xl";
        case "xl":
            return "w-full max-w-4xl";
        case "full":
            return "w-full h-full max-w-none";
        default:
            return "w-96 md:w-auto";
    }
};


export default GModalWrapper;