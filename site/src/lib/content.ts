import type { Locale, PageKey } from "./routes";

export const pricingPlans: Record<
  Locale,
  Array<{ name: string; price: string; cadence: string; scope: string; contactPath: string; enterprise?: boolean }>
> = {
  en: [
    {
      name: "Commercial Evaluation",
      price: "EUR 490",
      cadence: "one-time",
      scope: "30 days, 1 tenant, 1 deployment, non-production evaluation.",
      contactPath: "/contact/",
    },
    {
      name: "Starter",
      price: "EUR 9,900",
      cadence: "per year",
      scope: "1 tenant, 1 production deployment, up to 2 non-prod environments, 1-2 production surfaces.",
      contactPath: "/contact/",
    },
    {
      name: "Growth",
      price: "EUR 24,900",
      cadence: "per year",
      scope: "Up to 3 tenants and/or deployments, up to 6 production recommendation surfaces.",
      contactPath: "/contact/",
    },
    {
      name: "Enterprise",
      price: "From EUR 60,000",
      cadence: "per year",
      scope: "Custom scope for OEM, regulated deployments, multi-region HA, custom SLA, or legal/security terms.",
      contactPath: "/contact/",
      enterprise: true,
    },
  ],
  fi: [
    {
      name: "Kaupallinen arviointi",
      price: "EUR 490",
      cadence: "kertamaksu",
      scope: "30 päivää, 1 tenant, 1 käyttöönotto, ei-tuotannollinen arviointi.",
      contactPath: "/fi/yhteys/",
    },
    {
      name: "Starter",
      price: "EUR 9,900",
      cadence: "vuodessa",
      scope: "1 tenant, 1 tuotantokäyttöönotto, enintään 2 ei-tuotantoympäristöä ja 1-2 tuotannon suosittelupintaa.",
      contactPath: "/fi/yhteys/",
    },
    {
      name: "Growth",
      price: "EUR 24,900",
      cadence: "vuodessa",
      scope: "Enintään 3 tenantia ja/tai käyttöönottoa sekä enintään 6 tuotannon suosittelupintaa.",
      contactPath: "/fi/yhteys/",
    },
    {
      name: "Enterprise",
      price: "Alkaen EUR 60,000",
      cadence: "vuodessa",
      scope: "Mukautettu laajuus OEM-, säädeltyihin, multi-region HA-, SLA- tai sopimus- ja tietoturvatarpeisiin.",
      contactPath: "/fi/yhteys/",
      enterprise: true,
    },
  ],
};

export const fixedServices: Record<Locale, Array<[string, string, string]>> = {
  en: [
    ["Pilot Integration Review", "EUR 5,000", "Review one scoped pilot integration and evaluation readiness."],
    ["Production Readiness Package", "EUR 12,500", "Review rollback, hardening, observability, and cutover risks."],
    ["Security / Procurement Review", "EUR 5,000", "Support procurement review and hardening checklist navigation."],
  ],
  fi: [
    ["Pilotti-integraation katsaus", "EUR 5,000", "Yhden rajatun pilotti-integraation ja arviointivalmiuden katsaus."],
    ["Tuotantovalmiuden paketti", "EUR 12,500", "Palautuksen, kovennuksen, havainnoinnin ja cutover-riskien katsaus."],
    ["Tietoturva- ja hankintakatsaus", "EUR 5,000", "Tuki hankintakatsaukseen ja kovennuslistan läpikäyntiin."],
  ],
};

export const localized: Record<
  Locale,
  {
    ctaPrimary: string;
    ctaSecondary: string;
    contact: string;
    documentation: string;
    pages: Record<PageKey, { title: string; description: string; eyebrow: string; heading: string; intro: string }>;
  }
> = {
  en: {
    ctaPrimary: "Start an evaluation",
    ctaSecondary: "Read the docs",
    contact: "Contact",
    documentation: "Documentation",
    pages: {
      home: {
        title: "RecSys | Auditable recommendation system for evaluated rollouts",
        description:
          "RecSys is a self-hosted recommendation system suite with deterministic serving, offline evaluation, versioned artifacts, and rollback-ready operations.",
        eyebrow: "Self-hosted recommendation infrastructure",
        heading: "Recommendations you can evaluate, ship, and roll back.",
        intro:
          "RecSys gives technical and commercial teams a concrete path from first recommendation request to credible evaluation evidence and controlled production rollout.",
      },
      pricing: {
        title: "RecSys Pricing | Commercial evaluation and production plans",
        description:
          "Published RecSys commercial pricing anchors for evaluation, Starter, Growth, Enterprise, and fixed-scope review packages.",
        eyebrow: "Pricing",
        heading: "Clear commercial anchors for evaluation and rollout.",
        intro: "Start with a short evaluation, then move into a production scope that matches tenants, deployments, and surfaces.",
      },
      security: {
        title: "RecSys Security | Self-hosted recommendation system posture",
        description:
          "Security posture for RecSys: self-hosted deployment, pseudonymous identifiers, auth modes, tenancy, admin controls, audit logging, and limits.",
        eyebrow: "Security",
        heading: "Built for operator-controlled deployments.",
        intro:
          "RecSys is usually self-hosted, so your team controls infrastructure, secrets, network policy, backups, and retention. The product keeps the review surface explicit.",
      },
      evaluation: {
        title: "RecSys Evaluation | Ship recommendation changes with evidence",
        description:
          "Use RecSys evaluation workflows to validate exposure and outcome joins, guardrails, offline gates, and ship/hold/rollback decisions.",
        eyebrow: "Evaluation",
        heading: "Turn recommendation changes into defensible decisions.",
        intro:
          "A recommendation rollout should not depend on hope. RecSys keeps request IDs, exposure logs, outcome joins, reports, and rollback levers in the same operating model.",
      },
      contact: {
        title: "Contact RecSys | Commercial evaluation and licensing",
        description:
          "Contact RecSys for commercial evaluation, licensing, procurement review, security questions, or technical documentation.",
        eyebrow: "Contact",
        heading: "Start with the right channel.",
        intro:
          "Use public issues for reproducible public questions, private channels for commercial discussion, and the security reporting path for vulnerabilities.",
      },
      documentation: {
        title: "RecSys Documentation | Technical docs and procurement packet",
        description:
          "Find RecSys technical documentation, developer quickstart, API reference, operations runbooks, pricing, licensing, and procurement guidance.",
        eyebrow: "Documentation",
        heading: "Technical depth when you need it.",
        intro:
          "The marketing site gives the overview. The technical documentation keeps the implementation details, API reference, runbooks, and procurement packet.",
      },
      blog: {
        title: "RecSys Blog | Recommendation system evaluation and rollout notes",
        description:
          "Practical notes on auditable recommendation rollouts, evaluation workflows, and self-hosted recommendation infrastructure.",
        eyebrow: "Blog",
        heading: "Practical notes for evaluated recommendation systems.",
        intro:
          "Short articles for teams moving from recommendation experiments to auditable production rollouts.",
      },
    },
  },
  fi: {
    ctaPrimary: "Aloita arviointi",
    ctaSecondary: "Lue dokumentaatio",
    contact: "Yhteys",
    documentation: "Dokumentaatio",
    pages: {
      home: {
        title: "RecSys | Auditoitava suosittelujärjestelmä hallittuihin julkaisuihin",
        description:
          "RecSys on itse ylläpidettävä suosittelujärjestelmä, jossa on deterministinen tarjoilu, offline-arviointi, versioidut artefaktit ja palautusvalmiit julkaisut.",
        eyebrow: "Itse ylläpidettävä suositteluinfrastruktuuri",
        heading: "Suositukset, jotka voi arvioida, julkaista ja palauttaa.",
        intro:
          "RecSys antaa teknisille ja kaupallisille tiimeille selkeän polun ensimmäisestä suosittelupyynnöstä arviointinäyttöön ja hallittuun tuotantojulkaisuun.",
      },
      pricing: {
        title: "RecSys-hinnoittelu | Arviointi- ja tuotantosuunnitelmat",
        description:
          "RecSysin kaupalliset hinnoitteluankkurit arviointiin, Starter-, Growth- ja Enterprise-käyttöön sekä kiinteisiin arviointipaketteihin.",
        eyebrow: "Hinnoittelu",
        heading: "Selkeät kaupalliset vaihtoehdot arviointiin ja käyttöönottoon.",
        intro:
          "Aloita rajatulla arvioinnilla ja siirry tuotantolaajuuteen, joka vastaa vuokralaisia, ympäristöjä ja suosittelupintoja.",
      },
      security: {
        title: "RecSys-tietoturva | Itse ylläpidettävän suosittelujärjestelmän malli",
        description:
          "RecSysin tietoturvamalli: itse ylläpidettävä käyttöönotto, pseudonyymit tunnisteet, tunnistautuminen, tenant-eristys, ylläpitotoiminnot ja auditointi.",
        eyebrow: "Tietoturva",
        heading: "Suunniteltu operaattorin hallitsemiin ympäristöihin.",
        intro:
          "RecSys on yleensä itse ylläpidettävä, joten tiimisi hallitsee infrastruktuuria, salaisuuksia, verkkoa, varmuuskopioita ja säilytysaikoja.",
      },
      evaluation: {
        title: "RecSys-arviointi | Julkaise suosittelumuutokset näytön perusteella",
        description:
          "RecSysin arviointipolku auttaa tarkistamaan exposure- ja outcome-liitokset, guardrail-metriikat, offline-portit ja julkaisu-, odotus- tai palautuspäätökset.",
        eyebrow: "Arviointi",
        heading: "Muunna suosittelumuutokset perustelluiksi päätöksiksi.",
        intro:
          "Suosittelujulkaisun ei pidä perustua toiveeseen. RecSys yhdistää request ID:t, exposure-lokit, outcome-liitokset, raportit ja palautuskeinot samaan toimintamalliin.",
      },
      contact: {
        title: "Ota yhteyttä RecSysiin | Kaupallinen arviointi ja lisensointi",
        description:
          "Ota yhteyttä RecSysiin kaupallista arviointia, lisensointia, hankintakatsausta, tietoturvakysymyksiä tai teknistä dokumentaatiota varten.",
        eyebrow: "Yhteys",
        heading: "Valitse oikea yhteyspolku.",
        intro:
          "Käytä julkisia issueita toistettaviin julkisiin kysymyksiin, yksityisiä kanavia kaupalliseen keskusteluun ja tietoturvareittiä haavoittuvuuksiin.",
      },
      documentation: {
        title: "RecSys-dokumentaatio | Tekniset ohjeet ja hankintapaketti",
        description:
          "Löydä RecSysin tekninen dokumentaatio, kehittäjän pika-aloitus, API-viite, operointiohjeet, hinnoittelu, lisensointi ja hankintaohjeet.",
        eyebrow: "Dokumentaatio",
        heading: "Teknistä syvyyttä silloin kun sitä tarvitaan.",
        intro:
          "Markkinointisivusto antaa kokonaiskuvan. Tekninen dokumentaatio sisältää toteutusohjeet, API-viitteen, runbookit ja hankintapaketin.",
      },
      blog: {
        title: "RecSys-blogi | Suosittelujärjestelmien arviointi ja julkaisut",
        description:
          "Käytännön kirjoituksia auditoitavista suosittelujulkaisuista, arviointipoluista ja itse ylläpidettävästä suositteluinfrastruktuurista.",
        eyebrow: "Blogi",
        heading: "Käytännön muistiinpanoja arvioitaviin suosittelujärjestelmiin.",
        intro:
          "Lyhyitä artikkeleita tiimeille, jotka siirtyvät suosittelukokeiluista auditoitaviin tuotantojulkaisuihin.",
      },
    },
  },
};
