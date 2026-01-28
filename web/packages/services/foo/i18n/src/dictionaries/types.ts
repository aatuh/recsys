export type Dictionary = {
  brand: {
    name: string;
    shortName: string;
    metaTitle: string;
    metaDescription: string;
  };
  common: {
    footer: {
      themeLabel: string;
      themeSystem: string;
      themeLight: string;
      themeDark: string;
      themeVariantLabel: string;
      languageLabel: string;
      languageEnglish: string;
      languageFinnish: string;
    };
  };
  app: {
    errors: {
      loadFoos: string;
      saveFoo: string;
      deleteFoo: string;
    };
  };
};
