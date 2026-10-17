export function plantTitle(plant: { readonly title: string; readonly id: string }): string {
  return plant.title || plant.id;
}
