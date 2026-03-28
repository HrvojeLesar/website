package zkillboard

import "github.com/HrvojeLesar/website/eve_online/esi"

type killmailConverter struct{}

var KillmailConverter killmailConverter

func (converter *killmailConverter) FromEsiAndCorporationKillmail(esiKillmail *esi.ESIKillmail, corpKillmail *CorporationKillmail) Killmail {
	if esiKillmail == nil {
		panic("Esikillmail must not be null")
	}
	if corpKillmail == nil {
		panic("CorpKillmail must not be null")
	}

	return Killmail{
		KillmailID: esiKillmail.KillmailID,
		Hash:       corpKillmail.Zkb.Hash,
		Esi:        *esiKillmail,
		Zkb:        corpKillmail.Zkb,
	}
}
