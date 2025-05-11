export interface ElementData {
  tier: number;
  imageLink: string;
  recipes: string[][];
}

export interface ElementsData {
  [elementName: string]: ElementData;
}