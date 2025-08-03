

interface PropsType {
    name?: string;
    onClick: () => void;
    disabled?: boolean;
}

export const AddButton: React.FC<PropsType> = (props: PropsType) => {
    return (
        <button
            onClick={props.onClick}
            disabled={props.disabled}
            className="bg-blue-600 text-white px-4 py-2 rounded-lg font-medium hover:bg-blue-700 transition-colors"
            >
            {props.name || "Add"}
        </button>
    );
}

export default AddButton;