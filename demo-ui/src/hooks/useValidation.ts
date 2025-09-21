import { useState, useCallback } from "react";
import type {
  ValidationRule,
  ValidationErrors,
  TouchedFields,
} from "../types/ui";

export function useValidation<T extends Record<string, any>>(
  initialValues: T,
  rules: { [K in keyof T]?: ValidationRule }
) {
  const [errors, setErrors] = useState<ValidationErrors>({});
  const [touched, setTouched] = useState<TouchedFields>({});

  const validateField = useCallback(
    (name: keyof T, value: any): string | null => {
      const rule = rules[name];
      if (!rule) return null;

      // Required validation
      if (rule.required && (!value || value === "")) {
        return "This field is required";
      }

      // Skip other validations if value is empty and not required
      if (!value || value === "") return null;

      // Min/Max validation for numbers
      if (typeof value === "number") {
        if (rule.min !== undefined && value < rule.min) {
          return `Value must be at least ${rule.min}`;
        }
        if (rule.max !== undefined && value > rule.max) {
          return `Value must be at most ${rule.max}`;
        }
      }

      // Min/Max length validation for strings
      if (typeof value === "string") {
        if (rule.minLength !== undefined && value.length < rule.minLength) {
          return `Must be at least ${rule.minLength} characters`;
        }
        if (rule.maxLength !== undefined && value.length > rule.maxLength) {
          return `Must be at most ${rule.maxLength} characters`;
        }
      }

      // Pattern validation
      if (
        rule.pattern &&
        typeof value === "string" &&
        !rule.pattern.test(value)
      ) {
        return "Invalid format";
      }

      // Custom validation
      if (rule.custom) {
        return rule.custom(value);
      }

      return null;
    },
    [rules]
  );

  const validateAll = useCallback(
    (values: T): ValidationErrors => {
      const newErrors: ValidationErrors = {};

      for (const [name, value] of Object.entries(values)) {
        const error = validateField(name as keyof T, value);
        if (error) {
          newErrors[name] = error;
        }
      }

      setErrors(newErrors);
      return newErrors;
    },
    [validateField]
  );

  const setFieldError = useCallback((name: keyof T, error: string | null) => {
    setErrors((prev) => {
      if (error) {
        return { ...prev, [name]: error };
      } else {
        const { [name]: _, ...rest } = prev;
        return rest;
      }
    });
  }, []);

  const setFieldTouched = useCallback(
    (name: keyof T, touched: boolean = true) => {
      setTouched((prev) => ({ ...prev, [name]: touched }));
    },
    []
  );

  const getFieldError = useCallback(
    (name: keyof T): string | undefined => {
      return errors[name] || undefined;
    },
    [errors]
  );

  const isFieldTouched = useCallback(
    (name: keyof T): boolean => {
      return touched[name] || false;
    },
    [touched]
  );

  return {
    errors,
    touched,
    validateField,
    validateAll,
    setFieldError,
    setFieldTouched,
    getFieldError,
    isFieldTouched,
    hasErrors: Object.keys(errors).length > 0,
  };
}
