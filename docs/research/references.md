# Design references: what makes each one premium

Written 2026-10-29 as input to the look-dev phase. The binding context is `docs/brief/01-DESIGN-DIRECTION.md`: the committed world (a night-time hydrographic instrument, a plaster relief model in a dark map room), the references table in §2 and the anti-slop gate in §10. Nothing here overrides the brief. Where a reference does something the brief bans, this document says "do not take".

## Method and limits

- All research was done on 2026-10-29 with web search and web fetch. Every source in the list at the end was fetched that day. Quoted phrases were checked against the fetched page text.
- My tools return page text, documentation, HTML strings and source code. They do not show images, video or WebGL output. **Every observation below therefore comes from text, documentation, code or a method write-up, not from looking at the graphic.** Where a source says nothing about an aspect (for example nullschool's typefaces), I say so and leave it out.
- reuters.com, graphics.reuters.com and nytimes.com refuse my fetcher, and washingtonpost.com returned HTTP 403. For Reuters and the NYT I used the makers' own making-of interviews, a conference abstract and a newsletter that analysed the pieces. I did not see any of their graphics.
- shadedrelief.com (Tom Patterson's site) sits behind a Cloudflare browser challenge, which I did not try to bypass. Patterson appears here through his peer-reviewed paper with Bernhard Jenny and through the ETH Zurich relief-shading site, which documents the same Swiss tradition.

Each entry has one paragraph on what makes the piece premium (typography, light, camera, restraint, annotation), then 1 to 3 bullets on how it applies to มหานที.

---

## Part 1. The five references named in the brief

### 1. earth.nullschool.net (Cameron Beccario)

The visual is the data, and restraint is a stated rule rather than an accident. Beccario told Data Stories that he does not want to "clutter the page with text or buttons or social network buttons" and that he would publish it "exactly the same way" today. The only permanent chrome is a small "earth" label that opens a menu. Its rows are plain labels (Date, Data, Source, Scale, Control), so the source and the colour scale are one tap away but never sit on top of the data. **Typography** is modest and functional: the about page credits M+ and Google Noto fonts, the README says the M+ fonts are subsetted to cut download size, and the settings let the reader choose a font size from "tiny" to "huge" and switch on "Bold Text". The sources do not describe faces or label placement beyond that. **Colour** comes from perceptual science rather than mood: the about page credits ColorBrewer, Kindlmann linear-luminance, MYCARTA and Dave Green's cubehelix scales, and Beccario's own summary of his role is "I just pick the colors". There is no **light** in the 3D sense. The **camera** is a choice of nine named projections (Orthographic, Patterson, Winkel Tripel and others) that the reader drags and zooms. The view state lives in the URL hash fragment, so every view is a deep link and the back button works. Craft shows in the layer stack: a D3 canvas for coastlines and rivers, a WebGL canvas for the colour field, a 2D canvas for the particles and an SVG layer on top for "a few annotations" and an anti-aliased globe edge. He added that edge because earlier techniques made it "look a bit ragged". **Annotation** is minimal and on demand. Honesty is part of the design: the about page says the data come from numerical models with no guarantee of accuracy, and Beccario notes that the moving particles are not physical, yet some viewers still read their speed literally.

**How it applies to มหานที**
- One overlay in one place. The national sentence, the measured time, the source and the legend form a single typographic block, and the expandable legend (brief §9) opens from it. At country zoom, nothing else floats over the relief.
- Draw the Thai border, the coastline and the rivers as their own anti-aliased vector pass (a line layer, or SVG projected from 3D), separate from the terrain mesh, so edges never look ragged. This is the same reason nullschool added its SVG edge.
- No animated flow along rivers. We measure stage, not velocity, and readers take particle speed literally. Motion belongs to the gauges and the timeline only.

### 2. Mini Tokyo 3D (GitHub user nagix)

What makes it premium is the sense of a live system behaving correctly on real geography. Per its documentation, over 2,200 trains run at once from real timetables and real-time delay data "with a very smooth animation even on smartphones". Lines use "the same line colors used in official route maps with appropriate offsets according to the zoom level", so riders keep the colour code they already know, and the spacing between parallel lines is solved again at each zoom instead of overlapping. **Camera:** clicking a train, aircraft or station starts tracking. There are eight named viewpoints ("Position only", "Back", "Top-back", "Front", "Top-front", "Helicopter", "Drone", "Bird"), and clicking the map ends tracking. When a tracked train goes underground, the display switches to Underground Mode on its own: "the map will turn dark" and overground objects become see-through. The mode follows the subject, not the other way round. **Restraint** shows in what it withholds. Eco Mode lengthens the update interval "to reduce power consumption and device heating" (the default is 60 frames per second), and Playback Mode says plainly that "no real-time information will be reflected". **Annotation** is structured. A train panel lists the railway, train type, destination, train number, previous and next stations with times, the delay and operation information, and "delays and schedule changes are highlighted in orange": one colour with one meaning. A clock at the top left shows the date, time and time zone of the displayed data (Japan Standard Time by default). The documentation does not describe typography or lighting, and I could not see the render. **Do not take** the "orange spheres" drawn around trains delayed by more than a minute. They are halos by another name, which the brief bans.

**How it applies to มหานที**
- Following a station is an authored state with an explicit exit. ใกล้ฉัน locks the camera to your station, and tapping empty relief or pressing Escape returns to the national view. Ship one authored landing framing, not a viewpoint menu.
- One colour, one meaning. The alert colour appears only where water is above bank-full: in the gauge, in the panel numbers and in the 72-hour chart. It never appears as decoration.
- Show the data time as persistently as Mini Tokyo 3D shows its clock. When the timeline is scrubbed away from now, the scrubbed time replaces the measured time in the same place and the same type, so a past state cannot be mistaken for the present.

### 3. Eduard Imhof and Swiss relief cartography

Swiss-style relief is premium because every tonal decision serves the reading of form, and the method is written down. **Light:** the ETH relief-shading site says that "normally the cartographer lets the light 'shine' on the terrain from the upper-left" (north-west). Southern light can cause relief inversion, where mountains read as valleys. The cartographer also bends the light locally so that it strikes ridges and valleys at about 45°, avoiding "dead" ridges where the lit and shaded slopes share the same grey. **Contrast is graded by elevation** as aerial perspective: "contrast gradually sharpens towards the highest peaks and softens towards the lowlands". Fridolin Becker introduced this in his 1889 Canton Glarus map. His listed principles include "No cast shadows" and a medium tone for valley bottoms instead of white. ETH's flat-areas page adds that white plains leave hills "tonally isolated". Imhof, Becker's successor at ETH, turned the restraint into a recipe. In his photomechanical process (around 1945) he derived six printing plates from one grey shading: a soft light-green copy, five hypsometric grades, and a yellow plate for sun-lit slopes in which "flat areas are masked to avoid yellow colouration in even lowlands". The last plates had increasing contrast and strengthened only the highest peaks (Jenny and Hurni 2006). For small-scale atlas maps, his colour scale "ranges with increasing elevation from a light grey-blue-green by light olive to brownish-reddish yellow tones, and finally to white". It follows the principle "higher is brighter" and is deliberately lighter and more natural than Peucker's scale, whose saturation rises with height. **Generalisation** does the rest: at small scales "an exact geometric reproduction of an area is neither possible nor necessary", and structuring lines (rivers, coastlines, ridgelines) carry the frame. Jenny et al. (2020) note that analytical shading still cannot match manual shading at removing unneeded detail, adjusting light locally and varying brightness to make large landforms stand out. Imhof applied the style at 1:500,000 to 1:30 million, which is our country scale. Typography, camera and annotation are outside these sources, which are about the relief layer.

**How it applies to มหานที**
- Shade the relief, but do not cast shadows. Use a north-west key light at an azimuth near 315°, with no shadow maps and no screen-space ambient occlusion. If look-dev screenshots show a "dead" ridge, adjust the light locally in a baked shade texture rather than moving the scene light.
- Grade hillshade contrast by elevation. Contrast is strongest on the northern ranges and the Tenasserim hills and softest on the central plain, so the plain where most gauges cluster becomes a quiet ground for data.
- Warm the sun-lit slopes slightly, as Imhof's yellow plate did, but mask that warmth out of flat land so the Chao Phraya plain reads as one calm surface.

### 4. Reuters and NYT 3D flood and terrain graphics

I could not open any Reuters or NYT graphic, so this paragraph rests on the makers' own accounts and on one analyst's description, not on how the pieces look. What they describe is editorial discipline applied to camera and annotation. On *Snow Fall* (NYT, 2012), Graham Roberts says the Cascades flyover "was given a nice slow movement" so that readers could "absorb the topography" before the story began. The avalanche simulation's pacing "was driven by data", and the skiers' paths were drawn by the reader's scrolling. Andrew Kueneman says that at times "a more flashy, loud, or visually spectacular transition or behavior was toned down or removed". The team decided early "not to offer the exact same experience across all browsers/devices". On *Greenland Is Melting Away*, Derek Watkins says the zoomed-out view is a designed map built from imagery and land-shape data ("it's more a map"). The hard part was making the cross-fades between imagery scales invisible, a planned moulin diagram was cut to keep the focus, and visuals can "introduce error" through details nobody reported. For *Notre-Dame* (2024), the NYT 3D team decided not to "create a photorealistic model" and went monochrome "to keep the focus on the geometry", accepting "a more restricted camera path". William B. Davis of the NYT, in a NACIS 2024 abstract, argues for static JSON, PNG and MP4 files over slippy-map tiles, which "often include unnecessary data" and can render inconsistently. At Reuters Graphics, Simon Scarr and Marco Hernandez say "If a story does not benefit from a graphic then we don't make one". They iterate until about "version 20" and treat accuracy as a point where "There can be no compromise". A newsletter analysing the 2023 Derna flood coverage describes Reuters' 3D views of the dam sites as "intermediary views between purely representational maps and photos". The sources mention font choices only in passing, so I cannot describe the typography.

**How it applies to มหานที**
- Labels appear only after the camera settles. During the ใกล้ฉัน flight only the relief and the gauges move. The station name, the river and the sentence arrive when the flight ends, using the 160 to 260 ms UI timing.
- Draw only what is reported. No inundation extents, no inferred water surfaces, no rivers or names we cannot source. The relief stays "more a map": designed, between map and photo, never photoreal.
- Ship precomputed static assets for the relief (a mesh or height texture plus a baked shade texture, at two resolutions) instead of streaming terrain tiles. The first frame is then the same on every device and nothing pops in.

### 5. Apple Weather map layers

I could not see the app, so this rests on Apple's documentation and on reviews. What the texts describe is restrained controls and motion that carries the forecast. The full-screen maps cover precipitation, temperature, air quality and wind, and "the type of weather map shown can change depending on the weather conditions for the selected location", so the app chooses the relevant layer before the user does. Apple's iOS 15 announcement frames the redesign as "data and layouts that change based on conditions, such as heavy rain". The **timeline** is not permanent chrome: in a precipitation or wind map you "tap near the bottom of the screen to show the timeline", with a 1h/12h switch. MacRumors describes a progress bar that you pause, or scrub by dragging its "progress dot". Layers switch from one overlay control, and location actions hide behind touch-and-hold. MacRumors says the maps draw on Dark Sky, which Apple acquired in 2020. Nightingale's essay on Dark Sky names that lineage: start from now at the user's exact location, keep values comparable at a glance (temperature bars "relative to each other"), and use rough "light", "medium" and "heavy" ranges instead of false precision. Apple's map guidelines state the rules a premium map layer follows. The "muted style" is "a desaturated version of the map" for "information-rich content that you want to stand out". Selection uses "an outline and color variation". Overlapping points cluster and expand as you zoom. Attribution sits "10 points above the lowest resting position" of a bottom card. The sources do not describe the typefaces, the lighting or the easing curves, so I make no claims about them.

**How it applies to มหานที**
- On phones the timeline stays hidden until asked for: one draggable handle plus play/pause, stepping between the hourly measurements. Gauges transition between measured values in 160 to 260 ms and never show interpolated in-between values as if they were measured.
- The relief uses the muted style: desaturated and lower in contrast than any gauge, so the only saturated pixels on screen are data.
- On phone, keep the ThaiWater credit and the measured time visible just above the bottom sheet's lowest snap point, so the source never disappears under the sheet.

---

## Part 2. Further pieces

### 6. Tom Patterson and Bernhard Jenny, cross-blended hypsometric tints (*Cartographic Perspectives* 69, 2011), with ETH's flat-area method

This is the clearest written account of why a quiet land tint looks expensive. The problem the authors set out: conventional tints put dark green in every lowland, so readers take elevation colour for vegetation. Their fix varies lowland colour by environment. "Warm humid regions are rich yellow-green, cool humid regions are blue-green, arid regions are khaki brown, and polar regions are cold gray", and the zones blend across broad, fuzzy boundaries generalised from climate maps, because detailed gridded climate data gave edges that were too abrupt. Colour value was tuned by eye, with stated reasons. An arid brown as dark as the greens resulted in "vast areas of the world becoming somber and uninteresting", while extra red stopped lowlands reading as lowlands, so they settled on a neutral brown slightly lighter than the greens. **Light** is applied lightly. The shaded relief has no grey in flat areas, so the tints stay pure. Shadowed slopes are semitransparent and darken only the highest crests, and the lit north-west slopes lighten the tint beneath them. **Restraint** has a hierarchy: the mid and high tints are the "visual bait", and the lowland colours carry a quieter message. ETH's flat-area guidance completes the method for computed shading. Analytical hillshade makes flats too dark, with "disconcerting random values", so the guidance blends in a bright flat-area tone by a slope-dependent weight, optionally only below a set elevation. Lower flats can sit slightly darker than higher tableland. Typography, camera and annotation are outside these sources.

**How it applies to มหานที**
- Thailand is warm and humid, so start the lowland tint from a muted yellow-green paddy at low lightness, and desaturate the neighbours further. Drop the red-brown highland "bait": in this product, the bait must be the gauges.
- Remove the noise from the central plain. Blend toward a flat tone wherever the slope is below a threshold and the elevation is below a limit, choose both by eye in look-dev, and record them in the vertical-exaggeration ADR.
- Shadows darken only the highest crests. The plain receives no shadow at all.

### 7. Daniel Huffman, "Towards less Blender-y relief" (2022), with Stamen's survey of "Blender maps" (Alan McConchie, 2022)

Stamen traces the Blender-relief genre mostly to Huffman's 2017 tutorial. In this post he names the look this project must avoid and gives the fix. "The Look" is "exaggerated terrain that almost looks cinematic", where "every peak is separated by a deep black chasm" and every small bump competes with the major landforms. He blames Blender's strength at realistic shadows for pulling mappers toward a default, and stresses that "there's nothing intrinsic to the Blender method that requires reliefs to look this way". His method has three parts. First, generalise the DEM before rendering, using mean, Gaussian or median filters or Patterson's resolution bumping (blending the smoothed and original elevations at roughly 9:1 or 8:2, judged by eye). Second, lower the vertical exaggeration, because heavy exaggeration makes vector features appear to float above the terrain and its shadows hide terrain. Third, clip or fade the coastline after blurring. Above all he keeps effects subtle: "it takes less than you think for the effect to work", and shaded relief "doesn't always need to be dramatic". McConchie adds the reader's side. In the more dramatic maps, alpine valleys become almost black ("Hopefully there's nothing important down there…"), hyperreal relief invites literal reading of exaggeration, many maps are "already starting to feel the same", and shadows that overwhelm the data turn the map into chart junk. His advice is to tone down the drama where people live. Typography and camera are outside both texts.

**How it applies to มหานที**
- Preprocess the DEM with smoothing plus resolution bumping. Choose the blend ratio from look-dev screenshots and record it in the same ADR as the vertical exaggeration (brief §4).
- Test the exaggeration against the vectors. At the chosen value in the 6 to 15× range, rivers, the border and gauge bases must sit on the displaced surface, with no visible float or gap at country zoom or station zoom.
- People live, and gauges stand, in valleys and on plains (the Ping, Wang, Yom and Nan valleys, the southern coast). No valley may get dark enough that a gauge or label loses contrast. Check the computed contrast of the gauge marks against the darkest relief value.

### 8. Sam Learner, River Runner (2021)

River Runner lets a raindrop fall anywhere and follows it downstream. The Pudding Cup 2021 judges wrote: "It's rare for a project with so few words to give rise to so many cultural, ecological, and geographical considerations." Its quality lies in camera engineering that the README and source code make visible. The author "needed to make all sorts of adjustments for elevation and bearing changes to prevent jerkiness/nausea". In `Map.svelte`, the path is smoothed "by averaging coordinates with their neighbors" to "reduce horizontal movement with bendy rivers". The camera altitude rises with the terrain elevation sampled along the path, with an extra multiplier for mountainous starts. Pitch is derived from the camera's altitude and its distance to the target. The run's duration is computed from the route length so that speed stays constant. The opening is a single `flyTo` that pitches the camera into the raindrop view. **Annotation** is sparse and structured. A navigation box lists the named river features along the route and the total length in km. Clicking a feature jumps the run to it, Escape exits, and gaps in the data are labelled honestly ("Unnamed River/Stream"). Light and typography are not described in the README.

**How it applies to มหานที**
- Fly the ใกล้ฉัน path along a smoothed route, with the camera height derived from the relief beneath it, so that a flight to Nan or Chumphon never dips into the hills. Pick the duration within the 1.2 to 2.0 s range from the distance travelled, not as a fixed value.
- When a station has no river name or no bank-full height in the data, say so in the sentence and the panel (for example "ไม่มีข้อมูลระดับตลิ่ง") rather than hiding the station or guessing.

### 9. NOAA National Weather Service river gauge map and hydrograph conventions

This is not a showpiece. It is the most fully worked-out public design for exactly our kind of data, and the baseline a premium gauge view has to meet. Per the NWS *Guide to Hydrologic Information on the Web* (September 2015), every gauge state has an explicit legend entry. Besides the ordered flood categories, these include "River or stream gauge is not operating", "Observations are not current" and "Flood category has not been defined". Symbol shape encodes data availability (squares mark short-term forecasts, circles short-term plus probabilistic forecasts, and diamonds observations only), so colour is not overloaded. The hydrograph shows the observed trace, a "vertical line [that] marks the current viewing time", the colour-coded flood categories with the "stage where each category starts", the "latest observed stage with time and date", and the peak values. The guide states that the "flood severity color scheme is consistent throughout these web pages", so the map icon and the chart band always agree. The map can be filtered by category ("only points in that flood category are displayed"). The guide does not cover typography, light or camera, and its own graphics are utilitarian. The lesson is its states and its consistency, not its look.

**How it applies to มหานที**
- 72-hour chart: the observed trace only (we do not forecast), the bank-full line in the same token as the gauge cap, a "now" rule labelled with the measured time, and the 72-hour peak labelled.
- Stale data, undefined bank-full and an offline station are three distinct designed states, each with its own copy. They are shown by shape or fill (hollow for stale, per the brief), not by yet another hue.
- Tapping a level in the expandable legend filters the relief to that level (for example only น้ำล้นตลิ่ง), as a secondary action.

### 10. The Pudding, *Human Terrain* (Matt Daniels, 2018)

*Human Terrain* is a 3D relief of world population density. I saw only its text, but the text still shows the structure. There is a guided sequence of camera flights (a "Fly to Kinshasa" button, 21 stops), each with a one-line editorial caption such as "Kinshasa is now bigger than Paris". A time toggle switches between 1975, 1990 and 2015, with a "Compare to 1990" view. The loading line is plain ("Fetching Population Count..."). An honest methods note says the density data is "visibly noisy" where census data is unreliable, names India and China, and points to the published caveats. The premium lesson is the pairing of each camera move with exactly one sentence. The text does not describe light, typography or colour. One caution: the piece maps a non-elevation variable as terrain. That is legitimate here, because population is the subject, but it is also the root of slop pattern 6 below.

**How it applies to มหานที**
- Pair every flight with one sentence. The national sentence and the station sentence are the only captions, and a flight without a sentence does not ship.
- The place names in the national sentence (for example "ลุ่มเจ้าพระยาตอนล่าง") can be links that fly the camera there, giving a guided reading without a tour UI. This is a secondary action, subject to the gate.
- The about page names where our data are thin or noisy (stations without a bank-full value, telemetry gaps), the way *Human Terrain* names its weak regions.

### Thailand and Southeast Asia: what I looked for and did not find

- I searched for Reuters, NYT, Washington Post, Bloomberg, South China Morning Post and Pudding pieces on Chao Phraya floods, Bangkok floods or subsidence, Thai river gauges and the Mekong in 3D. I found **no 3D terrain or river-gauge graphic about the Chao Phraya** from these outlets. Reuters, NYT and Washington Post pages block my tools and search indexes Reuters Graphics poorly, so this result means "not found", not "does not exist". Bloomberg search returned news articles (mostly on the September and October 2026 Bangkok floods) and a Mekong video, not graphics.
- Kontinentalist (Singapore) has map-led Mekong stories: *A River Drained* (Peiying Loh, 26 June 2020, with Open Development Mekong) and *Dam if you do, dam if you don't* (1 May 2018, "Scroll to explore", with a skippable map segment). Their text was fetchable but their visuals were not, so I have not analysed them as design references.
- Robert Szucs's river-basin maps are known to me only through a secondary description (Open Culture, 2017): Strahler stream order, where "the higher the stream order, the thicker the line", with basins coloured by catchment and made in QGIS. That supports the brief's rule of line width by river order. Colouring basins in rainbow hues would compete with our single alert colour, so take the width rule only.

---

## Part 3. 3D slop patterns observed in the wild

These go beyond the brief's 3D tells list. Several come from Chinese "big-screen" (数据大屏) dashboard write-ups. They are anti-references that I read only to name the patterns, in line with the brief's instruction not to draw inspiration from dashboard templates.

1. **The extruded boundary slab with gradient side walls.** *Seen:* a juejin.cn write-up of a "可视化大屏 · 3D 地图", in which every province outline is extruded "to a thickness, forming raised 3D slabs" (拉出一定厚度，变成一块块立起来的 3D 板) whose side walls carry a "cyan-blue gradient from bottom to top" (青蓝渐变). An "inner shadow" darkens the map's edges and lightens its centre. The ThreeMap component ships the same thing as "GeoJSON 驱动的行政区域 3D 拉伸". *Why it happens:* extrusion is the cheapest way to make flat GeoJSON look 3D, and the genre equates thickness with a "tech feel" (the article: "深蓝底色，撑科技感"). *Counter-move:* Thailand is a continuous DEM surface that carries on into the neighbours and recedes through aerial perspective. No plinth, no side walls, no cut edge, no vignette. The border is a fine line on the surface.

2. **Looping "liveness" effects: sweep shaders and ripple rings.** *Seen:* ThreeMap's "侧面扫光 Shader 动画" (a highlight band sweeping up the side walls) and its "扩散点" (spreading-point) layer. The big-screen article adds a column-base halo that "rotates and spreads" (旋转、扩散), floor-grid cells that light up at random, and a light band that "slowly sweeps" (慢慢扫过去的光带) across the header. *Why it happens:* motion is added to signal "live" when the data are not changing, and points have no hierarchy, so all of them animate. *Counter-move:* nothing moves unless the data or the user moves it, and `frameloop="demand"` enforces this. "Live" is said in words, by the measured time. Emphasis comes from level of detail and size, plus the brief's once-only change mark.

3. **The floating, bobbing object with orbiting decoration, and an intro that locks the user out.** *Seen:* the same article lists "地图整体轻轻上下浮" (the whole map gently floats up and down) and outer rings that turn at different speeds and directions. Its intro pushes the camera in from far away over about 2 s and locks mouse input meanwhile (这段时间锁住鼠标操作). ThreeMap ships a rotating base disc (底图旋转装饰) and an `autoRotate` option. This is distinct from the brief's camera wobble, because here the object moves. *Why it happens:* idle animation fills empty seconds, and intro reels are treated as content. *Counter-move:* the relief is a fixed survey model on a table, perfectly still when idle. No intro flight. Every flight, including ใกล้ฉัน, can be interrupted by any input.

4. **Mirror floors and specular "sunset" water.** *Seen:* ThreeMap's "镜面反射" mirror-reflection water surface. The three.js `webgl_shaders_ocean` example, often the starting point for "water" in 3D, combines a reflective normal-mapped `Water` plane, a `Sky` dome with turbidity 10, a sun at elevation 2 (just above the horizon) and ACES filmic tone mapping. *Why it happens:* reflective water is the stock way to say "water" in WebGL. *Counter-move:* the sea is matte, dark and still, with no specular highlight and no normal map. Flood water is never rendered as a surface. Water exists only inside the gauges.

5. **The satellite drape at near-horizon pitch.** *Seen:* the Mapbox GL JS "Add 3D terrain" example, which uses the `standard-satellite` style, `exaggeration: 1.5` and `pitch: 80`. *Why it happens:* the shortest path to a "3D map" is the library demo, and photo texture both competes with the data and implies photoreal truth. *Counter-move:* no imagery at all, only an authored tint and baked shade. Clamp the polar angle well short of the horizon, so sky and haze never take over the frame. Set the exact limit in look-dev.

6. **Data as height.** *Seen:* Stamen's survey records creators rendering population density, rainfall, snowfall and household income as terrain. deck.gl's `HexagonLayer` exposes `extruded` and `elevationScale` and defaults to a ColorBrewer 6-class YlOrRd ramp. *Why it happens:* once a heightfield pipeline exists, every variable becomes a mountain, and the library's default ramp ships with it. *Counter-move:* the relief displaces real elevation only. Never extrude provinces or basins by flood counts, and use no aggregation layers. One gauge per station, coloured only from our tokens.

7. **Rainbow ramps on continuous values.** *Seen:* Crameri and Shephard, interviewed by the University of Oslo's *Titan* (2021). A rainbow "adds artificial boundaries". Rendering Mars topography in rainbow "hides whole craters in the greenish coloured areas", and blue lowlands suggest an ocean ("There is not"). *Why it happens:* the rainbow reads as "rich data" to its maker, who does not notice the false edges, and it fails readers with colour-vision deficiency. *Counter-move:* use a lightness-ordered hypsometric tint (Imhof's "higher is brighter") and a lightness-ordered situation scale, verified with deuteranopia and protanopia simulations as the brief requires.

8. **The stock dark basemap under the data.** *Seen:* CARTO's basemap-styles README describes Dark Matter as a "dark gray map, good for polygon/line data", "meant to be used as background or basemap below data layers". *Why it happens:* a dark theme plus the need for context pulls in a stock basemap, which brings its own fonts, label hierarchy and road network. The product then speaks in someone else's typography. *Counter-move:* no third-party basemap. Every label on screen is ours, set in our Thai face, from our own gazetteer.

9. **Atmosphere defaults that bring banned items back.** *Seen:* the Mapbox style-spec fog defaults are `star-intensity` 0.35 at zoom 5 (fading to 0 by zoom 6), a `space-color` running from near-black blue `#010b19` at zoom 4 to `#367ab9` at zoom 7, and `high-color` `#245cdf`, a saturated blue. *Why it happens:* switching fog on with defaults silently adds stars and a blue sky band at exactly the zoom range of a country view. *Counter-move:* if any map library is used, set every atmosphere property explicitly from tokens, and add a test that asserts star intensity is 0.

10. **The bathtub flood plane.** *Seen:* atlas.co's sea-level tool "fills areas below a given height with water based on terrain data only", colours them "dark blue" and "does **not** model levees, dams, sediment accumulation, or other flood defences". The tool calls itself "a visualisation tool, not a hazard map". *Why it happens:* a rising blue plane is the easiest 3D image of a flood. *Counter-move:* in this product it would be a prediction or an implied extent that we do not measure, which the brief forbids. No inundation surfaces, ever. Water levels appear only as gauge fills.

11. **Flat "even" lighting that kills the relief.** *Seen:* the big-screen article's three-light rig. One light comes from directly overhead, so that hover does not light one half more than the other, and one "lights evenly by surface orientation, without falloff, so brightness is about the same everywhere" (按表面朝向均匀给光、不衰减，保证各处亮度差不多). Depth then comes only from the side-wall gradient. *Why it happens:* even light avoids awkward dark patches on extruded shapes, but it removes form. *Counter-move:* one north-west key light plus a soft fill (brief §4), graded by elevation per Imhof, so that form is read from the surface itself.

12. **The cinematic Blender relief.** *Seen:* named by Huffman ("every peak is separated by a deep black chasm") and by Stamen (valleys "almost black", maps "already starting to feel the same"). *Why it happens:* ray-traced cast shadows plus high exaggeration are the tool's default spectacle. *Counter-move:* Swiss style. No cast shadows, a generalised DEM, toned flats, and the least exaggeration that keeps the plain readable (see entries 3, 6 and 7).

---

## Part 4. Synthesis: principles for look-dev

Each principle traces to the references above and ends with a check that can be run against the gate screenshots.

1. **Data is the only saturated colour.** The relief is muted and desaturated, and the alert colour has a single meaning. *Traces to:* nullschool, Apple's muted map style, Mini Tokyo 3D's orange, Patterson and Jenny's "visual bait". *Check:* in the 1440×900 screenshot, no relief pixel has higher chroma (OKLCH) than the level-4 gauge colour.

2. **Carved, not rendered.** A north-west key light is graded by elevation, with no cast shadows, no shadow maps and no ambient-occlusion halos. Flats are toned and the DEM is generalised. Exaggeration is the least that makes the plain readable, and every vector sits on the surface. *Traces to:* Imhof and Becker via ETH, Jenny and Hurni, Patterson and Jenny, Huffman, Stamen. *Check:* the central plain shows no speckle, no valley is black, and at station zoom the gauge bases and rivers touch the surface.

3. **Text arrives when the camera rests.** Labels are HTML, in a fixed place, one sentence per move. *Traces to:* *Snow Fall*, *Greenland*, *Human Terrain*, River Runner. *Check:* a frame-by-frame capture of the ใกล้ฉัน flight shows no station label before the final frame.

4. **The camera has a subject and a smooth, interruptible path.** The route is smoothed, the altitude follows the terrain, the duration (1.2 to 2.0 s) scales with distance, and there is an explicit exit. *Traces to:* Mini Tokyo 3D, River Runner, *Snow Fall*'s slow flyover. *Check:* any pointer or key input stops a flight, Escape and a tap on empty relief return to the national view, and reduced motion cuts instead of flying.

5. **Motion only when the data or the user moves, and never implied physics.** *Traces to:* nullschool (particle speed read literally), Mini Tokyo 3D's Eco Mode, slop patterns 2 and 3. *Check:* after 10 s idle the renderer draws no frames, and no river carries a flow animation.

6. **Time and source travel with every number.** *Traces to:* the NWS hydrograph (latest stage "with time and date", a "now" line), Mini Tokyo 3D's clock, nullschool's Date and Source rows. *Check:* every screenshot that shows a measurement also shows วัดเมื่อ and the ThaiWater source.

7. **Chrome hides until asked.** One overlay, a legend that expands, and on phones a timeline that appears on demand. *Traces to:* nullschool, Apple Weather. *Check:* the first frame at 390×844 shows only the relief, the gauges, the wordmark, the sentence, ใกล้ฉัน and search.

8. **Ordered by lightness, redundant by shape.** The situation scale rises monotonically in lightness and saturation. Stale is hollow, and offline and undefined bank-full have distinct marks. *Traces to:* nullschool's perceptual scales, Crameri, the NWS symbol shapes. *Check:* greyscale, deuteranopia and protanopia renders keep the level order readable.

9. **Draw only what is reported.** No inundation, no inferred surfaces, no invented names, and gaps are labelled as gaps. *Traces to:* Watkins ("introduce error"), River Runner's "Unnamed River/Stream", atlas.co's own caveat, the brief's non-goals. *Check:* every visible mark maps to a ThaiWater record or a cited geodata source.

10. **Every device gets a designed version, built from precomputed assets.** *Traces to:* *Snow Fall* (not the same experience on every device), the NYT NACIS talk on static files over tiles, Mini Tokyo 3D's Eco Mode, the brief's 2D fallback. *Check:* the no-WebGL and reduced-motion screenshots look intended rather than degraded, and first paint does not wait on tile streaming.

---

## Sources (all accessed 2026-10-29)

**earth.nullschool.net**
- https://earth.nullschool.net/about.html (accessed 2026-10-29)
- https://earth.nullschool.net/ (page text only; accessed 2026-10-29)
- https://github.com/cambecc/earth (README; accessed 2026-10-29)
- https://archive.datastori.es/episode/137 (Data Stories, Moritz Stefaner interviews Cameron Beccario; transcript; accessed 2026-10-29)

**Mini Tokyo 3D**
- https://github.com/nagix/mini-tokyo-3d and https://raw.githubusercontent.com/nagix/mini-tokyo-3d/master/README.md (accessed 2026-10-29)
- https://minitokyo3d.com/docs/master/ (accessed 2026-10-29)
- https://minitokyo3d.com/docs/master/user-guide/display-modes.html (accessed 2026-10-29)
- https://minitokyo3d.com/docs/master/user-guide/screen-and-operations.html (accessed 2026-10-29)

**Imhof and Swiss relief shading**
- https://ikgrelief.ethz.ch/design/light-direction/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/design/aerial-perspective/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/design/flat-areas/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/design/small-scales/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/design/plan-oblique-relief/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/history/swiss-style/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/cartographers/imhof/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/colours/hypsometric/ (accessed 2026-10-29)
- https://ikgrelief.ethz.ch/analytical/tone-flat-areas/ (accessed 2026-10-29)
- Jenny, B. and Hurni, L. (2006), "Swiss-Style Colour Relief Shading Modulated by Elevation and by Exposure to Illumination", *The Cartographic Journal* 43(3): https://mail.colororacle.org/berniejenny/pdf/2006_JennyHurni_SwissStyleShading.pdf (accessed 2026-10-29)
- Jenny, B. et al. (2020), "Cartographic Relief Shading with Neural Networks": https://arxiv.org/abs/2010.01256 (abstract; accessed 2026-10-29)

**Reuters and NYT**
- https://source.opennews.org/articles/how-we-made-snow-fall/ (accessed 2026-10-29)
- https://www.storybench.org/how-the-nyts-derek-watkins-designed-greenland-is-melting-away/ (accessed 2026-10-29)
- https://www.storybench.org/how-the-new-york-times-used-space-and-sound-to-revive-notre-dame-after-the-fire/ (accessed 2026-10-29)
- https://nacis2024.sched.com/event/1fwlN/practical-cartography-day-morning-session-i-3d-terrain-and-basemaps (William B. Davis, NYT, abstract; accessed 2026-10-29)
- https://onlinejournalismblog.com/2020/12/16/striking-the-balance-between-graphic-design-and-data-journalism-design-is-a-conversation/ (Simon Scarr and Marco Hernandez, Reuters; accessed 2026-10-29)
- https://vds.beehiiv.com/p/visual-data-storytelling-week-charts-14f2 (newsletter on the Derna coverage; accessed 2026-10-29)
- https://snd.org/?p=983 (SND, 2018 World's Best-Designed, Digital; consulted, little design detail; accessed 2026-10-29)
- https://github.com/reuters-graphics (organisation page; no 3D or terrain tooling listed; accessed 2026-10-29)

**Apple Weather**
- https://support.apple.com/en-us/guide/iphone/iph0960651f5/ios ("View weather maps on iPhone"; accessed 2026-10-29)
- https://www.apple.com/newsroom/2021/06/ios-15-brings-powerful-new-features-to-stay-connected-focus-explore-and-more/ (accessed 2026-10-29)
- https://macrumors.com/how-to/use-apple-new-weather-maps-ios (Tim Hardwick, 29 August 2021; accessed 2026-10-29)
- https://nightingaledvs.com/dark-sky-weather-data-viz/ (Srini Kadamati, "A Eulogy for Dark Sky, a Data Visualization Masterpiece", 22 March 2023; accessed 2026-10-29)
- https://developer.apple.com/design/human-interface-guidelines/maps (read through its JSON endpoint https://developer.apple.com/tutorials/data/design/human-interface-guidelines/maps.json; accessed 2026-10-29)

**Further pieces**
- https://cartographicperspectives.org/index.php/journal/article/download/cp69-patterson-jenny/html?inline=1 (Patterson and Jenny 2011; accessed 2026-10-29)
- https://somethingaboutmaps.wordpress.com/2022/01/13/towards-less-blender-y-relief/ (Daniel Huffman; accessed 2026-10-29)
- https://stamen.com/shadows-on-maps-are-getting-a-lot-more-exciting-and-heres-why/ (Alan McConchie, 16 September 2022; accessed 2026-10-29)
- https://github.com/sdl60660/river-runner (README; accessed 2026-10-29)
- https://raw.githubusercontent.com/sdl60660/river-runner/main/src/components/Map.svelte and https://raw.githubusercontent.com/sdl60660/river-runner/main/src/components/NavigationInfo.svelte (source; accessed 2026-10-29)
- https://pudding.cool/process/pudding-cup-2021/ (accessed 2026-10-29)
- https://www.weather.gov/media/water/brochures/Guide_to_Hydrologic_Information_Brochure.pdf (NWS, September 2015; accessed 2026-10-29)
- https://pudding.cool/2018/10/city_3d/ (page text; accessed 2026-10-29)
- https://www.openculture.com/2017/06/river-basins-of-the-us-in-wonderful-rainbow-colours-a-data-visualization-to-behold.html (on Robert Szucs; accessed 2026-10-29)
- https://kontinentalist.com/stories/dam-if-you-do-dam-if-you-dont-hydropower-in-the-mekong (accessed 2026-10-29)
- https://cdn-images.kontinentalist.com/static-html/food-security-mekong-river-hydropower-dam-climate-change/index.html (accessed 2026-10-29)

**Slop patterns**
- https://juejin.cn/post/7651812531245940746 ("可视化大屏 · 3D 地图"; accessed 2026-10-29)
- https://juejin.cn/post/7626521779408044059 (ThreeMap component; accessed 2026-10-29)
- https://threejs.org/examples/webgl_shaders_ocean.html (source; accessed 2026-10-29)
- https://docs.mapbox.com/mapbox-gl-js/example/add-terrain/ (accessed 2026-10-29)
- https://docs.mapbox.com/style-spec/reference/fog/ (accessed 2026-10-29)
- https://deck.gl/docs/api-reference/aggregation-layers/hexagon-layer (accessed 2026-10-29)
- https://titan.uio.no/english/2021/improper-use-colours-can-distort-scientific-data (Eivind Torgersen, 19 February 2021; accessed 2026-10-29)
- https://github.com/CartoDB/basemap-styles (accessed 2026-10-29)
- https://atlas.co/docs/sea-level-rise/ (accessed 2026-10-29)

## Could not access (2026-10-29)

- reuters.com and graphics.reuters.com: the fetch tool refuses these domains. Not seen: the Reuters Derna explainer linked from the VDS newsletter (reuters.com/graphics/LIBYA-STORM/EXPLAINER/klvyzqebzpg/) and the Reuters plastics pieces linked from visualisingdata.com.
- nytimes.com: the search tool refuses the domain. Not seen: *Snow Fall*, *Greenland Is Melting Away*, *Notre-Dame*, or the Libya flooding map linked from the VDS newsletter.
- washingtonpost.com (HTTP 403), designweek.co.uk (403), bigthink.com (403), weforum.org (403), opendata.swiss (403), shadedrelief.com (Cloudflare challenge), web.archive.org (refused by the fetch tool), pmc.ncbi.nlm.nih.gov (reCAPTCHA) and nature.com (login redirect), so the Crameri paper is cited through the University of Oslo interview instead.
- minitokyo3d.com/docs/ returned 404. The versioned docs under /docs/master/ worked.
