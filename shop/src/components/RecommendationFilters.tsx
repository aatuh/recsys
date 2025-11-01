"use client";
import { useState } from "react";
import { RecommendationConstraints, buildTagConstraints } from "@/lib/recommendations/constraints";

interface RecommendationFiltersProps {
  onFiltersChange: (constraints: RecommendationConstraints) => void;
  availableBrands?: string[];
  availableCategories?: string[];
  maxPrice?: number;
}

export function RecommendationFilters({ 
  onFiltersChange, 
  availableBrands = [],
  availableCategories = [],
  maxPrice = 1000 
}: RecommendationFiltersProps) {
  const [selectedBrands, setSelectedBrands] = useState<string[]>([]);
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);
  const [priceRange, setPriceRange] = useState<[number, number]>([0, maxPrice]);

  const updateFilters = () => {
    const constraints = buildTagConstraints({
      brands: selectedBrands.length > 0 ? selectedBrands : undefined,
      categories: selectedCategories.length > 0 ? selectedCategories : undefined,
      priceRange: priceRange[0] > 0 || priceRange[1] < maxPrice ? priceRange : undefined,
    });
    onFiltersChange(constraints);
  };

  const handleBrandToggle = (brand: string) => {
    const newBrands = selectedBrands.includes(brand)
      ? selectedBrands.filter(b => b !== brand)
      : [...selectedBrands, brand];
    setSelectedBrands(newBrands);
    setTimeout(updateFilters, 0);
  };

  const handleCategoryToggle = (category: string) => {
    const newCategories = selectedCategories.includes(category)
      ? selectedCategories.filter(c => c !== category)
      : [...selectedCategories, category];
    setSelectedCategories(newCategories);
    setTimeout(updateFilters, 0);
  };

  const handlePriceChange = (index: number, value: number) => {
    const newRange: [number, number] = [...priceRange];
    newRange[index] = value;
    setPriceRange(newRange);
    setTimeout(updateFilters, 0);
  };

  const clearFilters = () => {
    setSelectedBrands([]);
    setSelectedCategories([]);
    setPriceRange([0, maxPrice]);
    onFiltersChange({});
  };

  return (
    <div className="space-y-4 p-4 border rounded">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold">Filter Recommendations</h3>
        <button
          onClick={clearFilters}
          className="text-xs text-blue-600 hover:text-blue-800"
        >
          Clear all
        </button>
      </div>

      {/* Price Range */}
      <div>
        <label className="text-xs font-medium text-gray-700">Price Range</label>
        <div className="flex items-center gap-2 mt-1">
          <input
            type="number"
            min="0"
            max={maxPrice}
            value={priceRange[0]}
            onChange={(e) => handlePriceChange(0, Number(e.target.value))}
            className="w-20 px-2 py-1 text-xs border rounded"
            placeholder="Min"
          />
          <span className="text-xs text-gray-500">to</span>
          <input
            type="number"
            min="0"
            max={maxPrice}
            value={priceRange[1]}
            onChange={(e) => handlePriceChange(1, Number(e.target.value))}
            className="w-20 px-2 py-1 text-xs border rounded"
            placeholder="Max"
          />
        </div>
      </div>

      {/* Brands */}
      {availableBrands.length > 0 && (
        <div>
          <label className="text-xs font-medium text-gray-700">Brands</label>
          <div className="mt-1 space-y-1">
            {availableBrands.map((brand) => (
              <label key={brand} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={selectedBrands.includes(brand)}
                  onChange={() => handleBrandToggle(brand)}
                  className="text-xs"
                />
                <span className="text-xs">{brand}</span>
              </label>
            ))}
          </div>
        </div>
      )}

      {/* Categories */}
      {availableCategories.length > 0 && (
        <div>
          <label className="text-xs font-medium text-gray-700">Categories</label>
          <div className="mt-1 space-y-1">
            {availableCategories.map((category) => (
              <label key={category} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={selectedCategories.includes(category)}
                  onChange={() => handleCategoryToggle(category)}
                  className="text-xs"
                />
                <span className="text-xs">{category}</span>
              </label>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
