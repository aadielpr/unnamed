# Research: image resolution tiers for responsive serving

> Resolves wayfinder ticket [**Research: image resolution tiers for responsive serving**](https://github.com/aadielpr/unnamed/issues/4) (child of map [#1](https://github.com/aadielpr/unnamed/issues/1)).
> AFK research. Findings below cite primary sources only.
> **Scope:** this refines a *later, isolated migration of the upload pipeline*. It does **not** change the locked MVP (DECISIONS.md #6: thumbnail ~400px JPEG q75 + display ~2048px JPEG q80, skip raw originals).

## Question recap

How do professional image-heavy platforms handle responsive multi-resolution serving? Levels & dimensions, formats, naming/path conventions, `srcset` usage, pre-generate-vs-on-the-fly, and how all this refines the MVP's 2-version approach.

---

## 1. Two professional serving models

Real platforms pick one of two architectures — and many run both.

**(A) Pre-generate fixed variants at upload time.** The server produces a small set of named sizes (e.g. `thumb`, `small`, `medium`, `large`) and stores each in object storage. Serving is plain static file delivery. This is the WordPress / Smush / many-CDN pattern, and is what the MVP already does (2 versions).

**(B) On-the-fly transform at the edge / via an image API.** Store **one** master image and derive any size/format on request via URL params; the result is **cached at the edge** after the first request. Canonical products:

- **Cloudflare Images transformations** — URL form `https://<ZONE>/cdn-cgi/image/<OPTIONS>/<SOURCE-IMAGE>`; on a cache miss Cloudflare fetches the original, applies `width`/`format`/`quality`, caches the result, and serves it. "The original image is also cached to speed up future transformations." ([Cloudflare — Transformations overview](https://developers.cloudflare.com/images/optimization/transformations/overview/))
- **Imgix rendering API** — query-param API (`w=`, `dpr=`, `auto=format`, `q=`) over a single source. ([Imgix Rendering API overview](https://docs.imgix.com/apis/rendering/overview))
- **AWS CloudFront on-demand image resizing** — Lambda@Edge / CloudFront Functions resize an S3 original per request. ([AWS CloudFront docs](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/on-demand-image-resizing.html))

**Cost tradeoff:** pre-generate pays in *storage* (every asset × N sizes) and in *reprocessing* when you add a size; on-the-fly pays in *transform compute* (per unique variant, often per-month-metered) but lets you add sizes/formats later with **zero migration**. Cloudflare explicitly bills "the first request for each unique version within a calendar month" as one unique transformation. ([Cloudflare overview](https://developers.cloudflare.com/images/optimization/transformations/overview/))

## 2. How many levels, at what dimensions

There is **no universal standard**; dimensions are chosen from the largest CSS box the image will render in, multiplied by max DPR (commonly 2x). Two reference points are primary-source:

- **Cloudflare's own `width=auto` default breakpoints: `320, 768, 960, 1200` px**, snapping up to the smallest breakpoint ≥ detected viewport. ([Cloudflare — Make responsive images](https://developers.cloudflare.com/images/optimization/make-responsive-images/))
- The Cloudflare srcset examples in the same doc use width descriptors **320 / 640 / 960 / 1280 / 2560w** — a typical "a few widths covering phones → 4K".

For an event-photo gallery whose **max display is ~2048 long edge** (DECISIONS.md #6), a professional 3-tier set would be something like:

| Tier | long edge | use |
|------|----------|-----|
| thumb | ~256–320 px | grid/preview tiles (square-crop optional) |
| medium | ~640–960 px | phone detail / card view |
| display | ~2048 px | zoom / desktop full view |

…emitted via `srcset` width descriptors so the browser picks per device. (A pure 2-tier MVP — 400 + 2048 — is fine to ship; a 3-tier set is the natural refinement.)

## 3. Formats: JPEG vs AVIF/WebP

Facts, primary-sourced:

- **Output formats Cloudflare can serve:** PNG, JPEG, GIF, WebP, SVG, **AVIF**. ([Cloudflare — Limits and formats](https://developers.cloudflare.com/images/get-started/limits/))
- **`format=auto`** automatically serves "the most efficient format that the requesting browser supports"; this is the **default** for Cloudflare hosted images. **AVIF** is "an order of magnitude slower" to encode and Cloudflare falls back to WebP/JPEG if the image is too large to encode quickly. ([Cloudflare — Features: `format`](https://developers.cloudflare.com/images/optimization/features/))
- **AVIF edge limit: 1200px longest edge** at Cloudflare — so highest-resolution display tiers can't be AVIF via edge transform; they'd be WebP/JPEG. ([Cloudflare — Limits](https://developers.cloudflare.com/images/get-started/limits/))
- **MDN's own guidance:** "Formats like WebP and AVIF are recommended as they perform much better than PNG, JPEG, GIF for both still and animated images." ([MDN — `<img>` element, image formats](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/img))

Practical professional pattern: **`format=auto` (or pre-gen WebP + AVIF for the small tiers, JPEG fallback)** — modern browsers get the better codec, older browsers get JPEG, no markup forks. For the MVP (pre-gen, JPEG only) this is the main "free" upgrade to make whenever a transform layer exists.

## 4. Naming / path conventions and `srcset`

Two coexisting conventions:

**Path/param encoding of the variant.**
- Pre-generate + storage-key prefixes: `/low/x.jpg`, `/med/x.jpg`, `/high/x.jpg` (the convention the ticket names) — simple, portable, no special CDN feature needed.
- Edge-transform URL param: Cloudflare `/cdn-cgi/image/width=640,format=auto/<path>`; Imgix `?w=640&q=75&auto=format`. ([Cloudflare Features](https://developers.cloudflare.com/images/optimization/features/))
- Cloudflare Images hosted delivery: `https://imagedelivery.net/<ACCOUNT>/<IMAGE-ID>/<VARIANT-OR-OPTIONS>`; the `<VARIANT-OR-OPTIONS>` slot can be a **predefined named variant** *or* ad-hoc params. ([Cloudflare Features — URL breakdown](https://developers.cloudflare.com/images/optimization/features/))

**`srcset` + `sizes` in the markup.** Primary-source guidance (Cloudflare + MDN):

- For images that **scale with viewport** (`width:100%`/`50vw`), use **width descriptors `w`** plus `sizes` so the browser factors in *both* viewport width **and** device pixel ratio:
  ```html
  <img style="width:100%"
    srcset="/640.jpg 640w, /960.jpg 960w, /1280.jpg 1280w"
    src="/960.jpg" />
  ```
- For images with a **fixed CSS size**, use **density descriptors `2x`** to cover high-DPI screens (Cloudflare's own table: a 960px image on a 2x display looks blurry, so provide a 1920 version).
- `sizes` tells the browser the *layout* width, so it doesn't default to "image fills viewport". Example from Cloudflare docs:
  ```html
  <img style="max-width:640px"
    srcset=".../320.jpg 320w, .../640.jpg 640w, .../1280.jpg 1280w"
    sizes="(max-width:640px) 100vw, 640px" />
  ```
  So on a 2x screen > 640px the browser correctly picks the `1280w` entry.
- **`width=auto` is a no-markup alternative:** Cloudflare picks width from client hints (`sec-ch-viewport-width`) or user-agent detection fallback; one URL, no `srcset`. ([Cloudflare — Make responsive images](https://developers.cloudflare.com/images/optimization/make-responsive-images/))

## 5. Pre-generate at upload vs on-the-fly — the decision, sized for EventLens

| | Pre-generate (MVP today) | On-the-fly (edge transform) |
|---|---|---|
| Storage cost | N× storage (every asset × tiers) | 1 original |
| Compute cost | upload-time only | per unique variant, metered monthly |
| Add a new size later | re-process entire library | just change markup/params |
| Format flexibility | must pre-gen each codec | `format=auto`, free per-request |
| Needs special CDN feature | no — plain R2/S3 + CDN | yes — Cloudflare Images transform, Imgix, or CloudFront+Lambda |
| AVIF at high res | fine (you encode it) | constrained to 1200px at Cloudflare edge |

**Both are legitimate.** The pivot point is whether EventLens serves through **Cloudflare in front of R2** — at which point on-the-fly `format=auto` + `width=auto`/`srcset` becomes nearly free and removes the need to ever re-process. If the hosting topology is *not* Cloudflare-fronted, pre-generating 3 tiers (+ optional WebP) at upload is the simpler call.

## 6. Refinement of the MVP's 2-version approach

The locked MVP (`thumbnail ~400px JPEG q75` + `display ~2048px JPEG q80`, no raw) is **correct to ship**. The refinements below are for the later migration, not the MVP:

1. **Wire `srcset` + `sizes` in the SolidJS gallery now** (cheap, no backend change). With the existing 2 versions, phones stop downloading the 2048 display image. Highest ROI, zero infra. *Recommend doing this in the MVP build itself.* — check first, because DECISIONS.md #6 only specifies the *versions*, not the markup.
2. **Add a 3rd tier (~256–320px thumb) later** for denser grid previews — improves perceived load on a 200-guest event where many thumbnails load at once.
3. **Format → `format=auto` (WebP/AVIF) later**, once a transform layer exists. Either pre-gen WebP+AVIF at upload for the lower tiers, or switch to on-the-fly. This is the main bandwidth win (~25–35% smaller than JPEG at equal quality, per the codec guidance above) but is **meaningful only once serving is Cloudflare-fronted (or equivalent)** — see ticket #3.
4. **Keep raw originals out** (DECISIONS.md #6 — no raw in MVP). If the platform later moves to on-the-fly and billing tiers are introduced, store *one* full-res master and derive everything; that supersedes storing pre-gen tiers. Raw-full-as-premium remains out of scope for the MVP.

## 7. Dependencies / fog this surfaces

- The **format + architecture decision** (pre-gen with modern codecs vs on-the-fly) is **downstream of the hosting-topology ticket #3**. Don't ticket it until #3 resolves: the answer is materially different on Cloudflare-Fronted-R2 vs (e.g.) fly.io + S3.
- The "responsive-tier migration" itself is a future **task** ticket, not something to slice now — it's blocked by #3 *and* by the MVP going live (proving the event first). Leave it in the "later, isolated migration" bucket already noted in DECISIONS.md.

---

## Sources

- [Cloudflare Images — Transformations overview](https://developers.cloudflare.com/images/optimization/transformations/overview/) — on-cache-miss fetch+transform+cache, per-unique pricing.
- [Cloudflare Images — Make responsive images](https://developers.cloudflare.com/images/optimization/make-responsive-images/) — srcset/sizes patterns, `width=auto`, default breakpoints 320/768/960/1200, DPR explanation.
- [Cloudflare Images — Features (`format`, `width`)](https://developers.cloudflare.com/images/optimization/features/) — `format=auto` default, AVIF slower + fallback, `wbreakpoints`, URL interface.
- [Cloudflare Images — Limits and formats](https://developers.cloudflare.com/images/get-started/limits/) — supported input/output formats, AVIF 1200px edge limit.
- [MDN — `<img>` element](https://developer.mozilla.org/en-US/docs/Web/HTML/Element/img) — `srcset`/`sizes`/`<picture>` guidance; "WebP and AVIF recommended."
- [Imgix — Rendering API overview](https://docs.imgix.com/apis/rendering/overview) — query-param image API (`w`/`dpr`/`auto=format`).
- [AWS — On-demand image resizing with CloudFront](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/on-demand-image-resizing.html) — Lambda@Edge resize pattern.