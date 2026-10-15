import { describe, expect, it } from "vitest";
import { fitId, idCap, idRoom, markRoom, middleCut } from "./headerFit";

describe("header text fit", () => {
  it("cuts the middle of a long name and keeps its distinguishing tail", () => {
    expect(middleCut("production-line-a", 17)).toBe("production-line-a");
    const cut = middleCut("production-line-a", 11);
    expect(cut).toHaveLength(11);
    expect(cut.endsWith("line-a")).toBe(true);
    expect(cut).toContain("…");
  });

  it("sets a long id on one line at the overview and cuts it only at the 8 pixel floor", () => {
    const room = idRoom(0.35, null, "overview");
    const fitted = fitId("production-line-a", room, 0.35, idCap.overview);
    expect(fitted.size * 0.35).toBeCloseTo(8, 1);
    expect(fitted.text).toContain("…");
    expect(fitted.text.endsWith("line-a")).toBe(true);
    const short = fitId("compressor-1", room, 0.35, idCap.overview);
    expect(short.text).toBe("compressor-1");
    expect(short.size * 0.35).toBeGreaterThanOrEqual(8);
  });

  it("keeps the id at its wanted size when it fits", () => {
    expect(fitId("boiler-1", 206, 1, 25)).toEqual({ text: "boiler-1", size: 12.5 });
    expect(fitId("boiler-1", 206, 0.4, 25).size).toBe(25);
  });

  it("shrinks the id before it cuts it, down to 8 screen pixels", () => {
    const fitted = fitId("compressor-1", 146, 0.5, 25);
    expect(fitted.text).toBe("compressor-1");
    expect(fitted.size).toBeLessThan(25);
    expect(fitted.size * 0.5).toBeGreaterThanOrEqual(8);
  });

  it("cuts the middle once the floor is reached", () => {
    const fitted = fitId("production-line-a-extension", 120, 0.4, 25);
    expect(fitted.size).toBe(20);
    expect(fitted.text).toContain("…");
    expect(fitted.text.endsWith("n")).toBe(true);
  });

  it("reserves a narrow column for an icon mark and a capped one for words", () => {
    expect(markRoom(0.4, 0)).toBeLessThan(70);
    expect(markRoom(0.4, 40)).toBeLessThan(206 * 0.55 + 8.1);
    expect(idRoom(0.4, null)).toBe(206);
    expect(idRoom(0.4, 0)).toBeLessThan(206);
    expect(idRoom(0.3, null, "overview")).toBeLessThan(230);
    expect(idRoom(0.3, 0, "overview")).toBeLessThan(idRoom(0.3, null, "overview"));
  });
});
