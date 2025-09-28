import React, {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from "react";
import { type Theme, getThemeColors, generateThemeCSS } from "../ui/themes";
import { useStringStorage } from "../hooks/useStorage";

interface ThemeContextType {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  isDark: boolean;
  toggleTheme: () => void;
}

const ThemeContext = createContext<ThemeContextType | undefined>(undefined);

interface ThemeProviderProps {
  children: ReactNode;
  defaultTheme?: Theme;
}

export function ThemeProvider({
  children,
  defaultTheme = "auto",
}: ThemeProviderProps) {
  const { value: storedTheme, setValue: setStoredTheme } = useStringStorage(
    "theme",
    defaultTheme
  );
  const [theme, setTheme] = useState<Theme>(storedTheme as Theme);
  const [isDark, setIsDark] = useState(false);

  // Update theme colors when theme changes
  useEffect(() => {
    const colors = getThemeColors(theme);
    setIsDark(colors === getThemeColors("dark"));

    // Apply theme CSS to document
    const styleId = "theme-styles";
    let styleElement = document.getElementById(
      styleId
    ) as HTMLStyleElement | null;

    if (!styleElement) {
      styleElement = document.createElement("style");
      styleElement.id = styleId;
      document.head.appendChild(styleElement);
    }

    styleElement.textContent = generateThemeCSS(colors);

    // Update document class for additional styling
    document.documentElement.className = `theme-${theme}`;
    if (isDark) {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }, [theme, isDark]);

  // Listen for system theme changes when in auto mode
  useEffect(() => {
    if (theme !== "auto") return;

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = () => {
      const colors = getThemeColors("auto");
      setIsDark(colors === getThemeColors("dark"));

      const styleElement = document.getElementById(
        "theme-styles"
      ) as HTMLStyleElement | null;
      if (styleElement) {
        styleElement.textContent = generateThemeCSS(colors);
      }

      if (colors === getThemeColors("dark")) {
        document.documentElement.classList.add("dark");
      } else {
        document.documentElement.classList.remove("dark");
      }
    };

    mediaQuery.addEventListener("change", handleChange);
    return () => mediaQuery.removeEventListener("change", handleChange);
  }, [theme]);

  // Load theme from storage on mount
  useEffect(() => {
    if (storedTheme && ["light", "dark", "auto"].includes(storedTheme)) {
      setTheme(storedTheme as Theme);
    }
  }, [storedTheme]);

  // Save theme to storage when it changes
  useEffect(() => {
    setStoredTheme(theme);
  }, [theme, setStoredTheme]);

  const toggleTheme = () => {
    if (theme === "light") {
      setTheme("dark");
    } else if (theme === "dark") {
      setTheme("auto");
    } else {
      setTheme("light");
    }
  };

  const value: ThemeContextType = {
    theme,
    setTheme,
    isDark,
    toggleTheme,
  };

  return (
    <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
  );
}

export function useTheme(): ThemeContextType {
  const context = useContext(ThemeContext);
  if (context === undefined) {
    throw new Error("useTheme must be used within a ThemeProvider");
  }
  return context;
}
