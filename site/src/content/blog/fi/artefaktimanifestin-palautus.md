---
title: "Artefaktimanifestin palautus suosittelujärjestelmissä"
description: "Miten artefaktimanifestin palautus pitää suosittelumallit, featuret ja serving-artefaktit palautettavina ilman koko palvelun rollbackia."
language: "fi"
pubDate: "2026-05-09"
translationKey: "artifact-manifest-rollback"
tags: ["artefaktimanifesti", "rollback", "pipelinet"]
---

Artefaktimanifestin palautus erottaa suosittelujärjestelmän palauttamisen tilanteesta, jossa kaikki pitää julkaista
uudelleen yhden väärän mallin, feature-taulun tai sääntöpaketin vuoksi. Serving-järjestelmässä, jossa ranking ja data
muuttuvat usein, aktiivisen artefaktiversion pitää olla eksplisiittinen.

Palvelubinaari on vain yksi osa suosittelupolkua. Vastaus voi riippua myös mallitiedostoista, feature-snapshotista,
kandidaatti-indekseistä, config-säännöistä ja fallback-määrityksistä. Jos artefaktit liikkuvat ilman selkeää manifestia,
operoijat menettävät kyvyn selittää tai palauttaa tuotannon käyttäytymistä.

## Mitä artefaktimanifestin pitää tunnistaa

Artefaktimanifesti on osoitin versioituihin syötteisiin, joita serving-kerros saa käyttää. Sen pitää olla riittävän pieni
katselmoitavaksi ja riittävän tarkka tarjoillun vastauksen toistamiseen.

Vähintään manifestin pitäisi tunnistaa:

- malli- tai pisteytysartefaktien versiot
- kandidaatti- tai indeksiversiot
- feature-snapshotin tai feature-poiminnan versiot
- sääntö- ja config-versiot
- buildin tai pipeline-ajon metadata
- tuoreusaikaleimat
- validointitila

Manifestin ei tarvitse sisältää jokaista artefaktin sisältöä. Sen pitää osoittaa täsmällisiin versioihin, joista
serving-julkaisu koostuu.

## Miksi palvelun rollback ei riitä

API-julkaisun palautus voi olla väärä vipu, jos virhe tuli datasta tai artefakteista. Edellinen binaari voi yhä ladata
saman huonon manifestin. Uusi binaari voi olla terve, vaikka aktiivinen indeksi olisi vanhentunut.

Artefaktimanifestin palautus antaa operoijalle kapeamman korjauspolun. Sen sijaan, että palvelu julkaistaan uudelleen,
tiimi voi palauttaa viimeksi tunnetusti hyvän manifestiosoitteen, varmistaa että palvelu lukee sen ja seurata, että
guardrailit palaavat odotetulle tasolle.

Tämä parantaa myös häiriökatselmusta. Tiimi voi erottaa koodivirheet pipeline-virheistä, vanhentuneesta datasta ja
huonosta ranking-konfiguraatiosta.

## Artefaktimanifestin palautuspolku

Käytännöllinen palautuspolku on lyhyt ja eksplisiittinen:

1. Tunnista epäonnistunut julkaisu, vaikutettu tenant tai pinta ja aktiivinen manifestiversio.
2. Varmista guardrail tai arviointitulos, joka laukaisi palautuksen.
3. Valitse päätösjäljestä edellinen tunnetusti hyvä manifesti.
4. Siirrä aktiivinen manifestiosoitin takaisin siihen versioon.
5. Varmista valmius, tuoreus, vastauksen muoto ja exposure-lokitus.
6. Kirjaa palautettu versio ja rollbackin syy.

Tekniset dokumentit kuvaavat tämän työnkulun
[artefaktien ja pipelinejen oppaassa](/documentation/technical/artifacts-and-pipelines/) ja
[vanhentuneen artefaktimanifestin runbookissa](/documentation/technical/operations/runbooks/stale-artifact-manifest/).

## Backfillit tarvitsevat saman kurinalaisuuden

Backfillit ovat hyödyllisiä, mutta ne voivat tehdä tuotantotilasta vaikeamman ymmärtää. Jos backfill päivittää
featureita, indeksejä tai artefakteja, sen pitäisi tuottaa uusi manifestiehdokas eikä muuttaa aktiivisia
serving-syötteitä hiljaisesti.

Silloin arvioijilla ja operoijilla on selkeä päätöspiste: validoi uusi manifesti, julkaise se, pidä se odottamassa tai
palauta se.

## Yhdistä palautus arviointiin

Rollbackin pitää olla osa arviointisuunnitelmaa, ei hätätilanteessa keksitty käytäntö. Tiimin pitää tietää, mitkä
guardrailit merkitsevät, mikä manifestiversio on aktiivinen ja mikä edellinen versio on turvallinen ennen liikenteen
siirtoa.

Tuotetason työnkulku löytyy [arviointisivulta](/fi/arviointi/). Operatiiviseen korjaukseen käytä
[configin ja sääntöjen palautusrunbookia](/documentation/technical/operations/runbooks/rollback-config-rules/).
