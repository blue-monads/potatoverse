import React from 'react';
import { CapabilityDefinition, CapabilityOptionField } from '@/lib';
import { CapabilityOptionFieldInput } from './CapabilityOptionFieldInput';

interface CapabilityOptionsSectionProps {
    definition: CapabilityDefinition;
    formData: Record<string, any> | null | undefined;
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

    // Guard against null/undefined formData
    const safeFormData = formData || {};

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
                        value={safeFormData[field.key] !== undefined ? safeFormData[field.key] : field.default}
                        onChange={(value) => onFieldChange(field.key, value)}
                    />
                ))}
            </div>
        </div>
    );
};

