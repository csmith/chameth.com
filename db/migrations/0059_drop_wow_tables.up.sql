-- WoW character data now comes straight from the Ogre Stream API via the
-- cached data shortcodes, so the local mirror is no longer needed. The
-- rehosted portraits stay: the shortcodes keep serving the existing image
-- paths and update them in place on refresh.

DROP TABLE wow_achievements;
DROP TABLE wow_mythic_runs;
DROP TABLE wow_character_professions;
DROP TABLE wow_characters;
