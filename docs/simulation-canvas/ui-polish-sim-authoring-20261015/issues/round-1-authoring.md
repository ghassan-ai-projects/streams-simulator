# Round 1 - simulation authoring (UI comprehension test)

Result: blocked at step 1. Neither simulation was created. No plant was saved, no equipment was added, and no run was launched.
The tester stopped at the Home screen because there is no way to create a new plant from it.

Path: `http://127.0.0.1:8765`, single tab `seed`, viewport 1440x900 (reset to desktop afterwards). Not paired (no pairing screen).

## ISS-1-01 No "New plant" action on Home, so no plant can be created from the UI
- Severity: High
- Kind: not-working
- Where: Home / Workspace screen (breadcrumb "Workspace", heading "lab has no plants yet"). Header controls: "Streams Simulator" logo, "Keyboard", "System". The only creation controls are two "Try it" buttons under "Try an example" ("Chiller pilot", "Factory utilities").
- Steps:
  1. Open http://127.0.0.1:8765 in the seed tab.
  2. Look for a way to create a new plant (the brief says Home -> New plant).
  3. Use get_page_text, read_page (all and interactive) and find for "new plant", "new", "create" and "plant". Press Cmd+K.
- Observed: No "New plant" button, link or menu exists. The only way to add a plant is "Try it" on an example, which would create a copy of a fixed example (not a plant built in the editor). I did not click "Try it" because it would add an extra plant to the shared "lab" workspace and skip the build steps. Steps 1-6 of the brief could not start.
- Expected: A clear "New plant" (or "Create plant") action on the Home screen, with a prompt for id and title, then the editor opens on an empty plant.
- Suggestion: Add a primary "New plant" button next to the "Try an example" heading, and keep the examples as a secondary option.
- Status: fixed 66ec6f8
- Fix: the empty Home now has a primary "New plant" button above the examples; it opens a Name and Id form that creates a blank plant (the first shipped example is the hidden base, created with `empty`) and opens the editor. Unit and e2e (`empty-workspace.spec.ts`, own empty workspace in global-setup) tests added. The button stays disabled until the example list has loaded; `component new` has no base-less form, which would remove that dependency.

## ISS-1-02 Empty-state instructions point to files and folders, not to the UI
- Severity: Medium
- Kind: unclear
- Where: Home screen, numbered list under the examples.
- Steps:
  1. Open the Home screen with no plants.
  2. Read the first list item.
- Observed: "Try an example above, or copy a composition into compositions/ or a domain into domains/." The text names repository folders and implies hand-written files, which is a different route from the canvas.
- Expected: Instructions that match what the screen can do, for example "Create a plant" and "Open it and press Run".
- Suggestion: Replace the list with UI steps; move the file-based route to the documentation.
- Status: fixed 66ec6f8
- Fix: first step now reads "Create a new plant or try an example, then add its equipment in the editor."; no folder names on screen.

## ISS-1-03 "Keyboard" button changes state but shows nothing; "?" and Cmd+K do nothing
- Severity: Medium
- Kind: not-working
- Where: Header button "Keyboard (?)" (accessible name), right side of header with the keyboard icon.
- Steps:
  1. Click "Keyboard (?)" in the header.
  2. Take a screenshot.
  3. Press Shift+/ (the "?" key) on the page.
  4. Press Cmd+K.
  5. Read the page text after each step.
- Observed: The button shows a pressed/highlighted background after the click. No shortcut panel, dialog or list of shortcuts appears, and the page text does not change. The "?" key and Cmd+K also show nothing. A second click was not tried.
- Expected: A panel or dialog listing the keyboard shortcuts, with a visible way to close it. If there are no shortcuts on this screen, the button should not be there or should say so.
- Suggestion: Open a shortcuts dialog on click and on "?", and show the list on the editor screens where shortcuts exist.
- Status: won't-fix not reproducible: the Keyboard button opens the shortcuts dialog (four groups, Close button, Escape) on a real click and on a real "?" key press in Chrome at 1440x900. The tester's click came from a ref in the 1440 frame while the frame is 800x500, so it did not reach the button (the "pressed" look was hover). Cmd+K is not a feature of the canvas.
- Fix: none needed.

## ISS-1-04 Header label "System" does not say it is the theme control
- Severity: Low
- Kind: unclear
- Where: Header button with the monitor icon and visible label "System" (accessible name "Theme: follow the system").
- Steps:
  1. Read the header in the screenshot.
  2. Use read_page to read the accessible name of the button.
- Observed: The visible text is only "System". The meaning ("theme") appears only in the accessible name, so sighted users may read it as a system settings or status button.
- Expected: Visible text that says "Theme" or shows the current value, for example "Theme: System".
- Suggestion: Show the label "Theme: System" on screen, matching the accessible name.
- Status: fixed 2852fda
- Fix: the header button now reads "Theme: System", "Theme: Light" or "Theme: Dark"; its accessible name is "Theme: system|light|dark" so it matches the visible text.

## ISS-1-05 "lab" workspace name appears in the Home heading with no explanation
- Severity: Low
- Kind: unclear
- Where: Home heading "lab has no plants yet"; breadcrumb "Workspace"; tab title "lab · workspace · Streams Simulator".
- Steps:
  1. Open the Home screen.
  2. Read the heading and the breadcrumb.
- Observed: "lab" is used as a name for the current workspace, but the page never says what "lab" is or how to switch or rename it. The breadcrumb only says "Workspace".
- Expected: Either a label such as "Workspace: lab" with a way to change it, or no name in the heading.
- Suggestion: Show the workspace name next to the "Workspace" breadcrumb, with a switch control if there is more than one workspace.
- Status: fixed 20038db
- Fix: the empty heading reads "The lab workspace has no plants yet", the lead says plants are saved in the lab workspace folder, and the list header eyebrow says "Workspace folder". Switching workspaces is chosen when the bridge starts (`--workspace`), so no in-app switch exists (not built).

## Worked well
- The Home screen is clear and calm; the empty state explains what the tool is for in one sentence.
- The example cards have a name, a plain description and a labelled "Try it" button.
- Header buttons are reachable by keyboard-style accessible names (for example "Streams Simulator, workspace home").

## Simulation notes
- office-tower-hvac: not built. Blocked at ISS-1-01 (no New plant action). Area tree, equipment, instruments, actions: none.
- compressed-air-works: not built. Blocked at ISS-1-01 (no New plant action). Area tree, equipment, instruments, actions: none.
- Tester decision: I did not use "Try it" on the examples to get a plant into the editor. That would create an extra plant in the shared "lab" workspace and would not test the build path the brief asks for.

## Could not do
- Create office-tower-hvac: no New plant action on Home (ISS-1-01).
- Create compressed-air-works: same blocker.
- Add equipment, name instances, connect ports, organise areas, open the plant view, launch a run, schedule faults, use the scrubber and KPIs, try effectors or Run again, and reopen the run list: all depend on a plant existing in the editor, so none were reached.

## Tester notes (not app issues)
- The screenshot frame is 800x500 but clicks are in the 1440x900 frame. My first click at (712, 12) landed in empty header space; coordinates had to be taken from read_page refs. Using refs for clicks avoided this.
- No changes were made to the workspace, the repository, or git.

## Summary
- High: 1 (ISS-1-01)
- Medium: 2 (ISS-1-02, ISS-1-03)
- Low: 2 (ISS-1-04, ISS-1-05)
- Simulations completed: none
