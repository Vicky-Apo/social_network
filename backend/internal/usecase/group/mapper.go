package group

import domaingroup "social-network/backend/internal/domain/group"

func mapGroup(g domaingroup.Group) GroupDTO {
	return GroupDTO{
		ID:          g.ID,
		CreatorID:   g.CreatorID,
		Title:       g.Title,
		Description: g.Description,
		CreatedAt:   g.CreatedAt,
		UpdatedAt:   g.UpdatedAt,
	}
}

func mapGroups(groups []domaingroup.Group) []GroupDTO {
	out := make([]GroupDTO, 0, len(groups))
	for _, g := range groups {
		out = append(out, mapGroup(g))
	}
	return out
}

func mapMember(m domaingroup.GroupMember) GroupMemberDTO {
	return GroupMemberDTO{
		GroupID:  m.GroupID,
		UserID:   m.UserID,
		JoinedAt: m.JoinedAt,
	}
}

func mapMembers(members []domaingroup.GroupMember) []GroupMemberDTO {
	out := make([]GroupMemberDTO, 0, len(members))
	for _, m := range members {
		out = append(out, mapMember(m))
	}
	return out
}

func mapInvitation(inv domaingroup.GroupInvitation) GroupInvitationDTO {
	return GroupInvitationDTO{
		ID:        inv.ID,
		GroupID:   inv.GroupID,
		InviterID: inv.InviterID,
		InviteeID: inv.InviteeID,
		CreatedAt: inv.CreatedAt,
		UpdatedAt: inv.UpdatedAt,
	}
}

func mapInvitations(invs []domaingroup.GroupInvitation) []GroupInvitationDTO {
	out := make([]GroupInvitationDTO, 0, len(invs))
	for _, inv := range invs {
		out = append(out, mapInvitation(inv))
	}
	return out
}

func mapJoinRequest(req domaingroup.GroupJoinRequest) GroupJoinRequestDTO {
	return GroupJoinRequestDTO{
		ID:        req.ID,
		GroupID:   req.GroupID,
		UserID:    req.UserID,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}
}

func mapJoinRequests(reqs []domaingroup.GroupJoinRequest) []GroupJoinRequestDTO {
	out := make([]GroupJoinRequestDTO, 0, len(reqs))
	for _, req := range reqs {
		out = append(out, mapJoinRequest(req))
	}
	return out
}

func mapEvent(ev domaingroup.GroupEvent) GroupEventDTO {
	return GroupEventDTO{
		ID:          ev.ID,
		GroupID:     ev.GroupID,
		CreatorID:   ev.CreatorID,
		Title:       ev.Title,
		Description: ev.Description,
		EventTime:   ev.EventTime,
		CreatedAt:   ev.CreatedAt,
		UpdatedAt:   ev.UpdatedAt,
	}
}

func mapEvents(events []domaingroup.GroupEvent) []GroupEventDTO {
	out := make([]GroupEventDTO, 0, len(events))
	for _, ev := range events {
		out = append(out, mapEvent(ev))
	}
	return out
}
