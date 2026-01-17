"use client";
import { CapabilityDefinition } from "@/lib";


interface PropTypes {
    definations: CapabilityDefinition[]
    onSelect: (definition: CapabilityDefinition) => void
}

const CapabilityPicker = (props: PropTypes) => {
    const { definations, onSelect } = props;

    if (!definations || definations.length === 0) {
        return (
            <div className="text-center py-8 text-gray-500">
                No capability types available
            </div>
        );
    }

    return (
        <div className="flex flex-col gap-3">
            {definations.map((definition) => (
                <button
                    key={definition.name}
                    onClick={() => onSelect(definition)}
                    className="flex items-center gap-3 p-4 bg-white border border-gray-200 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-all duration-200 text-left group"
                >
                    {definition.icon && (
                        <div className="flex-shrink-0 w-10 h-10 flex items-center justify-center text-gray-600 group-hover:text-blue-600">
                            <span
                                dangerouslySetInnerHTML={{ __html: definition.icon }}
                                className="text-xl"
                            />
                        </div>
                    )}
                    <div className="flex-1 min-w-0">
                        <div className="font-medium text-gray-900 group-hover:text-blue-900">
                            {definition.name}
                        </div>
                        {definition.option_fields && definition.option_fields.length > 0 && (
                            <div className="text-xs text-gray-500 mt-1">
                                {definition.option_fields.length} option{definition.option_fields.length !== 1 ? 's' : ''}
                            </div>
                        )}
                    </div>
                </button>
            ))}
        </div>
    );
}

export default CapabilityPicker;