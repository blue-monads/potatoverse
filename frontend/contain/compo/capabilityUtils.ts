import { CapabilityDefinition, CapabilityOptionField } from '@/lib';

/**
 * Builds options object from form data and capability definition.
 * @param definition - The capability definition containing option fields
 * @param formData - The form data containing user input
 * @param includeDefaults - Whether to include default values for fields not in formData
 * @returns Options object ready to be saved
 */
export function buildCapabilityOptions(
    definition: CapabilityDefinition | undefined,
    formData: Record<string, any>,
    includeDefaults: boolean = false
): Record<string, any> | undefined {
    if (!definition) {
        return undefined;
    }

    const options: Record<string, any> = {};
    
    definition.option_fields.forEach(field => {
        if (formData[field.key] !== undefined) {
            options[field.key] = formData[field.key];
        } else if (includeDefaults && field.default !== undefined) {
            options[field.key] = field.default;
        }
    });

    return Object.keys(options).length > 0 ? options : undefined;
}

