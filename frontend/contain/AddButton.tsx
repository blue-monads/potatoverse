

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
            className="btn btn-base preset-filled"
            >
            {props.name || "Add"}
        </button>
    );
}

export default AddButton;