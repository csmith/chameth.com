---
title: Mix Composables and Views
group: Jetpack Compose
---

### Using a View in Compose

```kotlin
@Composable
fun Thing(modifier: Modifier = Modifier) {
    AndroidView(
        modifier = modifier,
        factory = { context ->
            // Create the view
            MyView(context)
        },
        update = { view ->
            // Optionally, configure the view once it's been inflated
            view.someProperty = myState.someField
        }
    )

}
```

### Using Compose in an Activity

Add the `androidx.activity:activity-compose` dependency, then:

```kotlin
class MyActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MaterialTheme {
                Text("Hack the planet!")
            }
        }
    }
}
```

### Using Compose in layouts

```xml
<androidx.compose.ui.platform.ComposeView
    android:id="@+id/new_shiny_thing"
    android:layout_width="match_parent"
    android:layout_height="match_parent" />
```

```kotlin
binding.newShinyThing.apply {
    setViewCompositionStrategy(ViewCompositionStrategy.DisposeOnViewTreeLifecycleDestroyed)
    setContent {
        MaterialTheme {
            Text("Hack the planet!")
        }
    }
}
```

### Using Compose programmatically in Views

```kotlin
ComposeView(requireContext()).apply {
    setViewCompositionStrategy(ViewCompositionStrategy.DisposeOnViewTreeLifecycleDestroyed)
    id = R.id.some_compose_view_id
    setContent {
        MaterialTheme {
            Text("Hack the planet!")
        }
    }
}
```

Each view needs a unique ID for data persistence.