type NotificationMeta = Record<string, unknown> | undefined;

export type NotificationLike = {
  type: string;
  metadata?: NotificationMeta;
  entity_id?: number;
};

export const allowedNotificationTypes = new Set([
  "follow_request",
  "group_invitation",
  "group_join_request",
  "event_created",
]);

export function getNotificationTitle(type: string): string {
  switch (type) {
    case "follow_request":
      return "Follow request";
    case "group_invitation":
      return "Group invitation";
    case "group_join_request":
      return "Join request";
    case "event_created":
      return "New group event";
    default:
      return "Notification";
  }
}

export function getNotificationActorName(metadata?: NotificationMeta): string {
  const requester = metadata?.["requester_name"];
  if (typeof requester === "string" && requester.trim()) return requester;
  return "Someone";
}

export function getNotificationGroupName(metadata?: NotificationMeta): string {
  const groupName = metadata?.["group_name"];
  if (typeof groupName === "string" && groupName.trim()) return groupName;
  return "your group";
}

export function getNotificationBody(type: string, metadata?: NotificationMeta): string {
  switch (type) {
    case "follow_request":
      return `${getNotificationActorName(metadata)} sent you a follow request.`;
    case "group_invitation":
      return `${getNotificationActorName(metadata)} invited you to ${getNotificationGroupName(metadata)}.`;
    case "group_join_request":
      return `${getNotificationActorName(metadata)} requested to join ${getNotificationGroupName(metadata)}.`;
    case "event_created":
      return `New event in ${getNotificationGroupName(metadata)}.`;
    default:
      return "Notification update.";
  }
}

export function getNotificationHref(item: NotificationLike): string | null {
  switch (item.type) {
    case "follow_request":
      return "/follow-requests";
    case "group_invitation":
      return "/group-invitations";
    case "group_join_request": {
      const groupID = item.metadata?.["group_id"];
      if (typeof groupID === "number") {
        return `/groups/${groupID}/join-requests`;
      }
      return "/groups";
    }
    case "event_created":
      return item.entity_id ? `/events/${item.entity_id}` : "/events";
    default:
      return null;
  }
}
