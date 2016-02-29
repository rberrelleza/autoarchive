CREATE TABLE groupinfo
(
    groupId integer NOT NULL,
    oauthId character varying(500) NOT NULL,
    oauthSecret character varying(500) NOT NULL,
    Created date,
    CONSTRAINT groupinfo_pkey PRIMARY KEY (groupid)
)
WITH (OIDS=FALSE);
