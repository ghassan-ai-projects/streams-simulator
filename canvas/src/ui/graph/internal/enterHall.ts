export function enterHall(id: string, onSelect: (id: string | null) => void, onOpenHall?: (hall: string) => void) {
  if (onOpenHall === undefined) {
    onSelect(id);
    return;
  }
  onSelect(null);
  onOpenHall(id);
}
