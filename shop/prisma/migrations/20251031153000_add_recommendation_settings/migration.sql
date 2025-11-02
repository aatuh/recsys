CREATE TABLE "RecommendationSettings" (
  "id" INTEGER NOT NULL PRIMARY KEY,
  "blendAlpha" REAL NOT NULL,
  "blendBeta" REAL NOT NULL,
  "blendGamma" REAL NOT NULL,
  "popularityHalflifeDays" REAL NOT NULL,
  "covisWindowDays" REAL NOT NULL,
  "popularityFanout" INTEGER NOT NULL,
  "mmrLambda" REAL NOT NULL,
  "brandCap" INTEGER NOT NULL,
  "categoryCap" INTEGER NOT NULL,
  "ruleExcludeEvents" INTEGER NOT NULL,
  "purchasedWindowDays" REAL NOT NULL,
  "profileWindowDays" REAL NOT NULL,
  "profileTopN" INTEGER NOT NULL,
  "profileBoost" REAL NOT NULL,
  "excludeEventTypes" TEXT NOT NULL DEFAULT 'view,click,add',
  "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "updatedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER "RecommendationSettings_updatedAt"
AFTER UPDATE ON "RecommendationSettings"
FOR EACH ROW
BEGIN
  UPDATE "RecommendationSettings" SET "updatedAt" = CURRENT_TIMESTAMP WHERE "id" = OLD."id";
END;

INSERT INTO "RecommendationSettings" (
  "id",
  "blendAlpha",
  "blendBeta",
  "blendGamma",
  "popularityHalflifeDays",
  "covisWindowDays",
  "popularityFanout",
  "mmrLambda",
  "brandCap",
  "categoryCap",
  "ruleExcludeEvents",
  "purchasedWindowDays",
  "profileWindowDays",
  "profileTopN",
  "profileBoost",
  "excludeEventTypes"
) VALUES (
  1,
  0.25,
  0.35,
  0.40,
  4,
  28,
  500,
  0.3,
  2,
  3,
  1,
  180,
  30,
  64,
  0.7,
  'view,click,add'
);
