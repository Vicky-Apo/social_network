/**
 * Maps API error messages to user-friendly text.
 */
export function formatApiError(error: string | undefined | null): string {
  if (!error) return "";
  const lower = error.toLowerCase().trim();
  if (lower === "forbidden") {
    return "You don't have permission to perform this action.";
  }
  if (lower === "unauthorized") {
    return "Please sign in to continue.";
  }
  if (lower === "not found") {
    return "This item could not be found.";
  }
  return error;
}
