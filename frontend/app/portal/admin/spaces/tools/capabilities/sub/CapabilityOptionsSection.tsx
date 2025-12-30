import React from 'react';
import { CapabilityDefinition, CapabilityOptionField } from '@/lib';
import { CapabilityOptionFieldInput } from './CapabilityOptionFieldInput';

interface CapabilityOptionsSectionProps {
    definition: CapabilityDefinition;
    formData: Record<string, any>;
    onFieldChange: (key: string, value: any) => void;
    className?: string;
}

export const CapabilityOptionsSection = ({
    definition,
    formData,
    onFieldChange,
    className = ''
}: CapabilityOptionsSectionProps) => {
    if (!definition || definition.option_fields.length === 0) {
        return null;
    }

    return (
        <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
                Configuration Options
            </label>
            <div className={`space-y-3 border border-gray-200 rounded-lg p-4 ${className}`}>
                {definition.option_fields.map(field => (
                    <CapabilityOptionFieldInput
                        key={field.key}
                        field={field}
                        value={formData[field.key] !== undefined ? formData[field.key] : field.default}
                        onChange={(value) => onFieldChange(field.key, value)}
                    />
                ))}
            </div>
        </div>
    );
};

