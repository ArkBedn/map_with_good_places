import folium
from folium.plugins import Draw
import webbrowser

# Startowa lokalizacja (Kraków Rynek)
start_coords = [50.061947, 19.936856]

# Tworzymy mapę
m = folium.Map(location=start_coords, zoom_start=13)

# Dodajemy narzędzie do rysowania (Draw)
draw = Draw(
    export=True,
    filename='obszar.geojson',
    position='topleft',
    draw_options={
        'polyline': False,
        'polygon': True,
        'circle': True,
        'rectangle': True,
        'marker': True,
        'circlemarker': False
    },
    edit_options={'edit': True}
)
draw.add_to(m)

# Zapisujemy mapę do pliku HTML
file_name = 'interaktywna_mapa.html'
m.save(file_name)

# Automatycznie otwieramy mapę w przeglądarce (w nowym oknie)
webbrowser.open_new_tab(file_name)
